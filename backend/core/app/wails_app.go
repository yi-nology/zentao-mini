package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"chandao-mini/backend/core/config"
	"chandao-mini/backend/core/logger"
	"chandao-mini/backend/core/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// WailsApp Wails桌面应用
// 实现Application接口，提供Wails桌面应用模式
// 该应用在Wails框架下运行，同时启动HTTP服务器提供API服务
type WailsApp struct {
	config *AppConfig
	deps   *Dependencies
	server *http.Server
	router *gin.Engine
	ctx    context.Context
	cancel context.CancelFunc
}

// NewWailsApp 创建Wails应用实例
func NewWailsApp(config *AppConfig, deps *Dependencies) *WailsApp {
	return &WailsApp{
		config: config,
		deps:   deps,
	}
}

// Start 启动Wails应用（由Wails框架调用）
func (a *WailsApp) Start(ctx context.Context) error {
	a.ctx = ctx

	// 创建可取消的上下文
	ctxWithCancel, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// 初始化配置
	if err := config.Init("", "ZENTAO_MINI"); err != nil {
		log.Printf("Warning: failed to initialize config: %v", err)
	}
	cfg := config.Get()

	// 初始化日志
	if err := logger.Init(&cfg.Log); err != nil {
		log.Printf("Warning: failed to initialize logger: %v", err)
	}

	// 在goroutine中启动后端服务
	go func() {
		// 设置路由
		a.router = routes.SetupRouterWithHandlers(
			a.deps.InitService,
			a.deps.ZentaoClient,
			a.deps.Handlers,
		)

		// 前端路由处理 - 所有非API请求都返回index.html
		a.router.NoRoute(func(c *gin.Context) {
			// 尝试从文件系统加载index.html
			indexPath := "./frontend/dist/index.html"
			if _, err := os.Stat(indexPath); os.IsNotExist(err) {
				// 如果dist目录不存在，使用public目录作为备选
				indexPath = "./frontend/public/index.html"
			}
			c.File(indexPath)
		})

		// 静态文件服务 - 提供前端资源
		a.router.Static("/assets", "./frontend/dist/assets")

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
			ReadTimeout:  120 * time.Second,
			WriteTimeout: 120 * time.Second,
		}

		logger.Info("Wails backend starting",
			zap.String("name", a.Name()),
			zap.String("port", port),
			zap.String("zentao_server", a.config.ZentaoServer),
		)

		// 启动MCP服务
		a.deps.Handlers.GetMCPHandler().Start()

		// 在goroutine中启动服务器
		go func() {
			if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("Failed to start server", zap.Error(err))
			}
		}()

		// 等待上下文取消
		<-ctxWithCancel.Done()
		logger.Info("Wails backend shutting down", zap.String("name", a.Name()))

		// 优雅关闭服务器
		if err := a.server.Shutdown(context.Background()); err != nil {
			logger.Error("Server forced to shutdown", zap.Error(err))
		}

		logger.Info("Wails backend stopped", zap.String("name", a.Name()))
	}()

	return nil
}

// Stop 停止Wails应用
func (a *WailsApp) Stop(ctx context.Context) error {
	// 取消上下文，停止后端服务
	if a.cancel != nil {
		a.cancel()
	}
	log.Printf("%s backend service stopped", a.Name())
	return nil
}

// Name 返回应用名称
func (a *WailsApp) Name() string {
	return "Wails-Desktop"
}

// Greet 示例方法 - Wails绑定方法
func (a *WailsApp) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
