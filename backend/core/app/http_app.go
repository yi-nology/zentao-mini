package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"chandao-mini/backend/core/logger"
	"chandao-mini/backend/core/routes"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPApp HTTP服务器应用
// 实现Application接口，提供独立的HTTP服务器模式
type HTTPApp struct {
	config    *AppConfig
	deps      *Dependencies
	server    *http.Server
	router    *gin.Engine
	ctx       context.Context
	cancel    context.CancelFunc
	embedMode bool // 是否使用嵌入的静态资源
}

// NewHTTPApp 创建HTTP应用实例
func NewHTTPApp(config *AppConfig, deps *Dependencies) *HTTPApp {
	return &HTTPApp{
		config:    config,
		deps:      deps,
		embedMode: config.StaticPath != "",
	}
}

// Start 启动HTTP服务器
func (a *HTTPApp) Start(ctx context.Context) error {
	a.ctx, a.cancel = context.WithCancel(ctx)

	// 设置路由
	a.router = routes.SetupRouterWithHandlers(
		a.deps.InitService,
		a.deps.ZentaoClient,
		a.deps.Handlers,
	)

	// 配置静态文件服务
	a.setupStaticFiles()

	// 获取端口
	port := a.config.Port
	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "12345"
		}
	}

	// 创建HTTP服务器
	a.server = &http.Server{
		Addr:         ":" + port,
		Handler:      a.router,
		ReadTimeout:  a.getReadTimeout(),
		WriteTimeout: a.getWriteTimeout(),
	}

	// 启动MCP服务
	a.deps.Handlers.GetMCPHandler().Start()

	// 在goroutine中启动服务器
	go func() {
		logger.Info("HTTP server starting",
			zap.String("name", a.Name()),
			zap.String("port", port),
			zap.String("zentao_server", a.config.ZentaoServer),
		)

		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	return nil
}

// getReadTimeout 获取读超时时间
func (a *HTTPApp) getReadTimeout() time.Duration {
	// 默认120秒
	return 120 * time.Second
}

// getWriteTimeout 获取写超时时间
func (a *HTTPApp) getWriteTimeout() time.Duration {
	// 默认120秒
	return 120 * time.Second
}

// Stop 停止HTTP服务器
func (a *HTTPApp) Stop(ctx context.Context) error {
	if a.cancel != nil {
		a.cancel()
	}

	if a.server != nil {
		logger.Info("HTTP server shutting down", zap.String("name", a.Name()))
		if err := a.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
		logger.Info("HTTP server stopped", zap.String("name", a.Name()))
	}

	return nil
}

// Name 返回应用名称
func (a *HTTPApp) Name() string {
	return "HTTP-Server"
}

// GetRouter 获取路由器（用于外部访问）
func (a *HTTPApp) GetRouter() *gin.Engine {
	return a.router
}

// setupStaticFiles 配置静态文件服务
func (a *HTTPApp) setupStaticFiles() {
	// 静态文件目录
	staticRoot := "./frontend/dist"
	assetsDir := staticRoot + "/assets"

	// 检查 dist 目录是否存在
	if _, err := os.Stat(staticRoot); os.IsNotExist(err) {
		logger.Warn("Frontend dist directory not found, static files disabled")
		return
	}

	// 提供静态文件服务
	a.router.Static("/assets", assetsDir)
	a.router.StaticFile("/", staticRoot+"/index.html")

	// 前端路由处理 - 所有非API请求都返回index.html (SPA支持)
	a.router.NoRoute(func(c *gin.Context) {
		// 跳过 API 路径
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.Next()
			return
		}
		// 跳过 metrics 路径
		if c.Request.URL.Path == "/metrics" || c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		c.File(staticRoot + "/index.html")
	})

	logger.Info("Static files configured", zap.String("root", staticRoot))
}
