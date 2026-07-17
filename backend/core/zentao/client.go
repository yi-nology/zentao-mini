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
		c.mu.RLock()
		defer c.mu.RUnlock()
		tokenStr = c.token.Get()
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
