package zentao

import (
	"context"
	"fmt"
	neturl "net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/metrics"

	"github.com/yi-nology/common/biz/zentao"
	"go.uber.org/zap"
)

// Client 封装禅道 SDK 客户端，支持 Token 缓存
type Client struct {
	sdkClient   *zentao.Client
	account     string
	password    *SecureString
	server      string
	token       *SecureString
	tokenExpiry atomic.Int64
	mu          sync.RWMutex
	connected   atomic.Bool
	refreshing  atomic.Bool

	// session 模式相关字段。mode == AuthModeSession 时 session != nil，
	// 所有数据请求经 withSessionRetry 走 *.json 端点；否则走原有 Token 路径。
	mode    AuthMode
	realm   string
	session *SessionTransport
}

const tokenTTL = 23 * time.Hour

// NewClient 创建新的禅道客户端
func NewClient(server, account, password string) *Client {
	// server 为空时跳过 URL 规范化，避免生成 "http:/" 这样的无效地址；
	// 后续可通过 UpdateConfig 在用户上传配置后补全。
	if server != "" {
		if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
			server = "http://" + server
		}
		server = strings.TrimSuffix(server, "/")
	}

	sdkClient := zentao.NewClient(server)
	sdkClient.SetTimeout(120 * time.Second)
	client := &Client{
		sdkClient: sdkClient,
		account:   account,
		password:  NewSecureString(password),
		server:    server,
		token:     NewSecureString(""),
	}
	go client.startTokenRefreshTask()
	return client
}

// NewSessionClient 创建会话模式的禅道客户端。
// 登录走 /user-login.html（PHP 会话），数据访问走 *.json 端点。
// 适用于禁用 REST API 的禅道实例（如麒麟 pm.kylin.com）。
func NewSessionClient(server, account, password, realm string) *Client {
	if server != "" {
		if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
			server = "http://" + server
		}
		server = strings.TrimSuffix(server, "/")
	}
	client := &Client{
		account:  account,
		password: NewSecureString(password),
		server:   server,
		token:    NewSecureString(""),
		mode:     AuthModeSession,
		realm:    realm,
		session:  NewSessionTransport(server, account, password, realm),
	}
	// 会话模式下异步触发首次登录（与 Token 模式的异步 RefreshToken 行为对齐）。
	if server != "" && account != "" {
		go func() {
			if err := client.session.ensureLoggedIn(); err != nil {
				logger.Warn("异步会话登录失败", zap.Error(err))
				return
			}
			client.connected.Store(true)
		}()
	}
	go client.startSessionKeepalive()
	return client
}

// startSessionKeepalive 会话模式下的保活：每 30 分钟检查一次，
// 若会话过期则在下次请求时自动重新登录（ensureLoggedIn 已覆盖）。
// 这里仅负责把 connected 标志位同步成实际状态。
func (c *Client) startSessionKeepalive() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		if c.mode != AuthModeSession || c.session == nil {
			return
		}
		c.connected.Store(c.session.IsLoggedIn())
	}
}

func (c *Client) startTokenRefreshTask() {
	ticker := time.NewTicker(2 * time.Hour)
	defer ticker.Stop()

	for {
		<-ticker.C
		if c.isTokenExpired() {
			if _, err := c.RefreshToken(); err != nil {
				continue
			}
		}
	}
}

func (c *Client) isTokenExpired() bool {
	expiry := time.Unix(c.tokenExpiry.Load(), 0)
	return time.Now().After(expiry)
}

// IsTokenExpired 检查 Token 是否已过期
func (c *Client) IsTokenExpired() bool {
	return c.isTokenExpired()
}

func (c *Client) getToken() (string, error) {
	tokenStr := c.token.Get()
	if tokenStr != "" && !c.isTokenExpired() {
		metrics.RecordCacheHit("token")
		return tokenStr, nil
	}

	metrics.RecordCacheMiss("token")

	if !c.refreshing.CompareAndSwap(false, true) {
		// 另一个 goroutine 正在刷新 token：等待它完成（最多 10 秒），
		// 而不是立即失败。这避免了 dashboard 并发拉取时部分请求拿不到 token 的问题。
		for i := 0; i < 50; i++ {
			time.Sleep(200 * time.Millisecond)
			// 检查刷新是否已完成（refreshing 标志位被清回 false）
			if !c.refreshing.Load() {
				c.mu.RLock()
				tokenStr = c.token.Get()
				c.mu.RUnlock()
				if tokenStr != "" && !c.isTokenExpired() {
					return tokenStr, nil
				}
				// 刷新完了但 token 还是无效（可能刷新失败），跳出循环
				break
			}
		}
		// 最后兜底：再读一次 token
		c.mu.RLock()
		tokenStr = c.token.Get()
		c.mu.RUnlock()
		if tokenStr != "" && !c.isTokenExpired() {
			return tokenStr, nil
		}
		return "", fmt.Errorf("token 刷新进行中，请稍后重试")
	}
	defer c.refreshing.Store(false)

	start := time.Now()
	tokenStr, err := c.doRefreshToken()
	if err != nil {
		c.connected.Store(false)
		logger.Error("Failed to get token after retries", zap.Error(err))
		return "", err
	}

	c.connected.Store(true)
	metrics.RecordCacheOperation("token", "refresh", time.Since(start))
	return tokenStr, nil
}

// RefreshToken 强制刷新 Token
func (c *Client) RefreshToken() (string, error) {
	if !c.refreshing.CompareAndSwap(false, true) {
		return "", fmt.Errorf("token 刷新进行中")
	}
	defer c.refreshing.Store(false)

	start := time.Now()
	tokenStr, err := c.doRefreshToken()
	if err != nil {
		c.connected.Store(false)
		return "", err
	}

	c.connected.Store(true)
	metrics.RecordCacheOperation("token", "refresh", time.Since(start))
	return tokenStr, nil
}

func (c *Client) doRefreshToken() (string, error) {
	c.mu.RLock()
	account := c.account
	passwordStr := c.password.Get()
	sdk := c.sdkClient
	c.mu.RUnlock()

	start := time.Now()
	var token string
	var err error
	for i := 0; i < 3; i++ {
		token, err = sdk.GetToken(account, passwordStr)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.token.Set(token)
	c.tokenExpiry.Store(time.Now().Add(tokenTTL).Unix())
	c.sdkClient.SetToken(token)
	c.mu.Unlock()

	metrics.RecordTokenRefresh()
	logger.Info("Token refreshed successfully",
		zap.Duration("duration", time.Since(start)),
	)

	return token, nil
}

// UpdateConfig 更新客户端配置并异步刷新Token
func (c *Client) UpdateConfig(server, account, password string) error {
	if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
		server = "http://" + server
	}
	server = strings.TrimSuffix(server, "/")

	c.mu.Lock()
	c.server = server
	c.account = account
	c.password.Set(password)
	c.sdkClient = zentao.NewClient(server)
	c.sdkClient.SetTimeout(120 * time.Second)
	c.token.Set("")
	c.tokenExpiry.Store(0)
	GlobalCache.Clear()
	c.mu.Unlock()

	go func() {
		if _, err := c.RefreshToken(); err != nil {
			logger.Warn("异步刷新Token失败", zap.Error(err))
		}
	}()

	return nil
}

// UpdateSessionConfig 会话模式的热重载入口。同步触发一次登录验证凭据，
// 返回登录错误（成功返回 nil）。失败时 connected 标志为 false。
func (c *Client) UpdateSessionConfig(server, account, password, realm string) error {
	if server != "" {
		if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
			server = "http://" + server
		}
		server = strings.TrimSuffix(server, "/")
	}

	c.mu.Lock()
	c.server = server
	c.account = account
	c.password.Set(password)
	if realm != "" {
		c.realm = realm
	}
	c.mode = AuthModeSession
	if c.session == nil {
		c.session = NewSessionTransport(server, account, password, realm)
	} else {
		c.session.UpdateConfig(server, account, password, realm)
	}
	GlobalCache.Clear()
	c.mu.Unlock()

	// 同步登录：登录表单提交后需要立即知道凭据是否有效，
	// 不能像 Token 模式那样纯异步（否则前端拿不到明确的成败反馈）。
	if err := c.session.ensureLoggedIn(); err != nil {
		c.connected.Store(false)
		return err
	}
	c.connected.Store(true)
	return nil
}

func (c *Client) GetServer() string {
	return c.server
}

func (c *Client) IsConnected() bool {
	return c.connected.Load()
}

// GetAccount 获取当前登录用户的账号
func (c *Client) GetAccount() string {
	return c.account
}

// GetMode 返回当前认证模式（token / session）。
func (c *Client) GetMode() AuthMode {
	return c.mode
}

// GetRealm 返回会话模式的认证域（kydc / local），token 模式下为空。
func (c *Client) GetRealm() string {
	return c.realm
}

// IsSessionMode 是否会话模式。
func (c *Client) IsSessionMode() bool {
	return c.mode == AuthModeSession
}

// SessionLoginSync 同步执行一次会话登录。仅供 InitHandler.Login 在表单提交时
// 显式调用以获得同步的成败反馈。Token 模式下返回 ErrNotSessionMode。
func (c *Client) SessionLoginSync() error {
	if c.mode != AuthModeSession || c.session == nil {
		return ErrNotSessionMode
	}
	if err := c.session.ensureLoggedIn(); err != nil {
		c.connected.Store(false)
		return err
	}
	c.connected.Store(true)
	return nil
}

// ErrNotSessionMode 在 token 模式下调用会话专用方法时返回。
var ErrNotSessionMode = fmt.Errorf("当前客户端不是会话模式")

// doSessionJSON 在会话模式下请求 *.json 端点。鉴权失败（zentaosid 过期）时
// 强制重新登录并重试一次；权限拒绝等永久错误直接返回。
// path 形如 "/bug-browse-1029-0-all.json"，out 为内层 data 的目标结构指针。
func (c *Client) doSessionJSON(ctx context.Context, operation, path string, out interface{}) error {
	if c.mode != AuthModeSession || c.session == nil {
		return ErrNotSessionMode
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	err := c.session.DoJSON(ctx, path, out)
	if err == nil || !IsSessionAuthError(err) {
		return err
	}
	logger.Warn("Zentao session expired, re-login and retry",
		zap.String("operation", operation),
		zap.String("path", path),
		zap.Error(err),
	)
	// 强制重新登录：清空 cookie jar 和过期标记，确保 ensureLoggedIn 真正执行 login()
	// （否则 IsLoggedIn 会因 cookie 仍在而误判有效，陷入重试死循环）。
	c.session.invalidateForRelogin()
	if loginErr := c.session.ensureLoggedIn(); loginErr != nil {
		return fmt.Errorf("%s失败，重新登录失败: %w，原始错误: %v", operation, loginErr, err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return c.session.DoJSON(ctx, path, out)
}

// doSessionPost 在会话模式下提交写操作表单。鉴权失效时重新登录后重试一次。
func (c *Client) doSessionPost(ctx context.Context, operation, path string, form neturl.Values) error {
	if c.mode != AuthModeSession || c.session == nil {
		return ErrNotSessionMode
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	err := c.session.PostForm(ctx, path, form)
	if err == nil || !IsSessionAuthError(err) {
		return err
	}
	logger.Warn("Zentao session expired on write, re-login and retry",
		zap.String("operation", operation),
		zap.String("path", path),
		zap.Error(err),
	)
	c.session.invalidateForRelogin()
	if loginErr := c.session.ensureLoggedIn(); loginErr != nil {
		return fmt.Errorf("%s失败，重新登录失败: %w，原始错误: %v", operation, loginErr, err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return c.session.PostForm(ctx, path, form)
}

func (c *Client) withTokenRetry(operation string, call func(*zentao.Client) error) error {
	if _, err := c.getToken(); err != nil {
		return err
	}

	c.mu.RLock()
	sdk := c.sdkClient
	c.mu.RUnlock()

	err := call(sdk)
	if err == nil || !isAuthError(err) {
		return err
	}

	logger.Warn("Zentao token rejected, refreshing and retrying",
		zap.String("operation", operation),
		zap.Error(err),
	)

	if _, refreshErr := c.RefreshToken(); refreshErr != nil {
		return fmt.Errorf("%s失败，刷新Token失败: %w，原始错误: %v", operation, refreshErr, err)
	}

	c.mu.RLock()
	sdk = c.sdkClient
	c.mu.RUnlock()
	return call(sdk)
}

func (c *Client) withTokenRetryContext(ctx context.Context, operation string, call func(*zentao.Client) error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, err := c.getToken(); err != nil {
		return err
	}

	c.mu.RLock()
	sdk := c.sdkClient
	c.mu.RUnlock()

	err := call(sdk)
	if err == nil || !isAuthError(err) {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	logger.Warn("Zentao token rejected, refreshing and retrying",
		zap.String("operation", operation),
		zap.Error(err),
	)

	if _, refreshErr := c.RefreshToken(); refreshErr != nil {
		return fmt.Errorf("%s失败，刷新Token失败: %w，原始错误: %v", operation, refreshErr, err)
	}

	c.mu.RLock()
	sdk = c.sdkClient
	c.mu.RUnlock()
	return call(sdk)
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	if zentao.IsAuthError(err) {
		return true
	}
	msg := strings.ToLower(err.Error())
	authMarkers := []string{
		"状态码: 401",
		"status code: 401",
		"unauthorized",
		"状态码: 403",
		"status code: 403",
		"forbidden",
		"authentication required",
	}
	for _, marker := range authMarkers {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}
