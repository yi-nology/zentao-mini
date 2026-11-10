package zentao

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/metrics"

	"github.com/oekit/pkg/httpjson"
	"github.com/oekit/pkg/secretx"
	"github.com/yi-nology/common/biz/zentao"
	"go.uber.org/zap"
)

// Client 封装禅道 SDK 客户端，支持 Token 缓存
type Client struct {
	sdkClient   *zentao.Client
	account     string
	password    *secretx.SecureString
	server      string
	token       *secretx.SecureString
	tokenExpiry atomic.Int64
	mu          sync.RWMutex
	connected   atomic.Bool
	refreshing  atomic.Bool
	// refreshInterval / requestTimeout 配置面接进来的两个值（>0 恒成立，
	// NewClientWithConfig 已归一）。UpdateConfig 重建 sdkClient 时用后者。
	refreshInterval time.Duration
	requestTimeout  time.Duration
	webSession      struct {
		mu      sync.Mutex
		session webSession
	}
}

const (
	tokenTTL = 23 * time.Hour
	// 缺省值：配置未给（或给的非法）时用——与 config.go 的 SetDefault 同一口径
	//（token_refresh_interval=12h、request_timeout=120s）。
	defaultRefreshInterval = 12 * time.Hour
	defaultRequestTimeout  = 120 * time.Second
)

// NewClient 创建新的禅道客户端（缺省参数形态：刷新间隔 12h / 请求超时 120s）。
func NewClient(server, account, password string) *Client {
	return NewClientWithConfig(server, account, password, 0, 0)
}

// NewClientWithConfig 创建禅道客户端并把**配置面接上**：token 刷新间隔与
// 请求超时不再是硬编码——yaml 的 zentao.token_refresh_interval /
// zentao.request_timeout 从此生效（此前两个字段 viper 读入、setDefaults
// 校验，client 却写死 2h ticker 与 120s 超时：配置在撒谎）。
//
// 零/负值取缺省（refreshInterval=12h，requestTimeout=120s）。
func NewClientWithConfig(server, account, password string, refreshInterval, requestTimeout time.Duration) *Client {
	// server 为空时 NormalizeBaseURL 原样返回空串（「未配置」语义不被改写），
	// 非空时补 scheme/去尾斜杠——单源在 pkg/httpjson，此前与 UpdateConfig
	// 各自内联一份。后续可通过 UpdateConfig 在用户上传配置后补全。
	server = httpjson.NormalizeBaseURL(server)

	if refreshInterval <= 0 {
		refreshInterval = defaultRefreshInterval
	}
	if requestTimeout <= 0 {
		requestTimeout = defaultRequestTimeout
	}
	sdkClient := zentao.NewClient(server)
	sdkClient.SetTimeout(requestTimeout)
	client := &Client{
		sdkClient:       sdkClient,
		account:         account,
		password:        secretx.NewSecureString(password),
		server:          server,
		token:           secretx.NewSecureString(""),
		refreshInterval: refreshInterval,
		requestTimeout:  requestTimeout,
	}
	// 已配置完整凭据时启动即建立连接，
	// 避免首次打开心跳/仪表盘前 IsConnected() 一直为 false 造成"未连接"误报
	if server != "" && account != "" {
		go func() {
			// NewClient 可能在 logger/metrics 初始化前被调用（如 mcp 入口、单元测试），
			// 预连接失败或 panic 都不允许带崩进程。recover 对齐 health.go 的范式：
			// 吞可以，但要留痕——无日志的 recover 让「初始化路径在崩」变成永远无人知道的事。
			defer func() {
				if r := recover(); r != nil {
					logger.Error("启动预连接 panic（已吞，进程不受影响）", zap.Any("panic", r))
				}
			}()
			if _, err := client.RefreshToken(); err != nil {
				logger.Warn("启动时连接禅道失败", zap.Error(err))
			}
		}()
	}
	go client.startTokenRefreshTask()
	return client
}

func (c *Client) startTokenRefreshTask() {
	interval := c.refreshInterval
	if interval <= 0 {
		interval = defaultRefreshInterval
	}
	ticker := time.NewTicker(interval)
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

// wrapTokenError 把 token 获取失败的常见误配翻译成可操作的提示，
// 最典型的就是服务器启用了 https 却配了 http:// 地址：POST 会被 301 到
// HTML 页面，SDK 解析时报 "invalid character '<'"，用户完全看不出原因。
func wrapTokenError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "invalid character") {
		return fmt.Errorf("%w（禅道服务器返回了非 JSON 内容：常见原因是服务器启用了 https，但配置的是 http:// 地址，请检查服务器地址）", err)
	}
	return err
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
		return "", wrapTokenError(err)
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
	// 与 NewClientWithConfig 同源：单源 pkg/httpjson.NormalizeBaseURL
	// （空串原样返回，不再像旧内联版那样把空 server 规范成 "http:/"）
	server = httpjson.NormalizeBaseURL(server)

	c.mu.Lock()
	c.server = server
	c.account = account
	c.password.Set(password)
	c.sdkClient = zentao.NewClient(server)
	c.sdkClient.SetTimeout(c.requestTimeout)
	c.token.Set("")
	c.tokenExpiry.Store(0)
	c.mu.Unlock()

	// 清缓存放在临界区外：Clear 需要 cache 全局写锁，若在持有 client.mu 时调用，
	// 与"缓存加载协程持 cache 锁等待 client.mu 刷新 token"形成互相等待死锁
	GlobalCache.Clear()
	c.ResetWebSession()

	go func() {
		if _, err := c.RefreshToken(); err != nil {
			logger.Warn("异步刷新Token失败", zap.Error(err))
		}
	}()

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

// RefreshInterval token 刷新间隔（issue #30：供装配层断言配置真实接线，
// 此前 yaml 字段被静默丢弃无从观测）
func (c *Client) RefreshInterval() time.Duration {
	return c.refreshInterval
}

// RequestTimeout 请求超时（同上，配置面观测口）
func (c *Client) RequestTimeout() time.Duration {
	return c.requestTimeout
}

// refreshOrWait 刷新 Token；若另一协程正在刷新则等待其完成（最长 15 秒）而不是立即失败。
// 定时任务常在同一分钟同时触发（如工作日 9 点多个报告），Token 恰好过期时会并发 401
// → 同时抢刷新，输家若直接放弃整条推送就会失败（2026-10-08 每日bug 线上故障根因）。
func (c *Client) refreshOrWait(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.refreshing.CompareAndSwap(false, true) {
		defer c.refreshing.Store(false)
		_, err := c.doRefreshToken()
		return err
	}
	for i := 0; i < 150; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		time.Sleep(100 * time.Millisecond)
		if !c.refreshing.Load() {
			// 持锁者已完成刷新（无论成败），调用方用新 Token 重试
			return nil
		}
	}
	return fmt.Errorf("等待其他协程完成Token刷新超时（15s）")
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

	if refreshErr := c.refreshOrWait(nil); refreshErr != nil {
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

	if refreshErr := c.refreshOrWait(ctx); refreshErr != nil {
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
