package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"chandao-mini/backend/core/initialization"
	"chandao-mini/backend/core/routes"
	"chandao-mini/backend/core/zentao"

	"github.com/gin-gonic/gin"
)

// App struct
type App struct {
	ctx    context.Context
	r      *gin.Engine
	cancel context.CancelFunc
	server *http.Server
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 创建可取消的上下文
	ctxWithCancel, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	// 启动后端服务
	go func() {
		// 初始化服务
		initService := initialization.NewInitService(
			os.Getenv("AUTH_CONFIG_PATH"),
			os.Getenv("AUTH_DB_PATH"),
			os.Getenv("ENCRYPTION_KEY"),
		)

		// 加载禅道配置
		zentaoServer, zentaoAccount, zentaoPassword := initialization.LoadZentaoConfig(initService)

		// 创建禅道客户端
		zentaoClient := zentao.NewClient(zentaoServer, zentaoAccount, zentaoPassword)

		// 设置路由
		r := routes.SetupRouter(initService, zentaoClient)
		a.r = r

		// 获取端口
		port := os.Getenv("PORT")
		if port == "" {
			port = "12345"
		}

		// 创建 HTTP 服务器
		server := &http.Server{
			Addr:    ":" + port,
			Handler: r,
		}
		a.server = server

		log.Printf("Server starting on port %s", port)
		log.Printf("Zentao Server: %s", zentaoServer)
		log.Printf("Frontend available at http://localhost:%s", port)

		// 在 goroutine 中启动服务器
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start server: %v", err)
			}
		}()

		// 等待上下文取消
		<-ctxWithCancel.Done()
		log.Println("Shutting down backend server...")

		// 优雅关闭服务器
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}

		log.Println("Backend server exited properly")
	}()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// shutdown is called when the app exits
func (a *App) shutdown(ctx context.Context) {
	// 取消上下文，停止后端服务
	if a.cancel != nil {
		a.cancel()
	}
	log.Println("Backend service stopped")
}
