package main

import (
	"context"
	"embed"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"chandao-mini/backend/core/app"
	"chandao-mini/backend/core/config"
	"chandao-mini/backend/core/logger"
	"chandao-mini/backend/core/metrics"

	"github.com/gin-gonic/gin"
)

// 编译时通过 ldflags 注入的配置变量
// 使用方式: go build -ldflags "-X main.encryptionKey=xxx -X main.zentaoServer=xxx ..."
var (
	encryptionKey  = ""
	zentaoServer   = ""
	zentaoAccount  = ""
	zentaoPassword = ""
	authDBPath     = "./data/auth.db"
)

//go:embed static/*
var embeddedStaticFS embed.FS

// getFileSystem 返回嵌入的静态文件系统
func getFileSystem() http.FileSystem {
	// 从嵌入的文件系统中获取static目录
	staticFS, err := fs.Sub(embeddedStaticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(staticFS)
}

func main() {
	// 初始化日志
	if err := logger.Init(&config.LogConfig{
		Level:            "info",
		Format:           "console",
		OutputPath:       "",
		EnableCaller:     true,
		EnableStacktrace: false,
	}); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 初始化性能指标
	if err := metrics.Init(); err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}

	// 使用编译时注入的配置变量（可通过 ldflags 覆盖）
	appConfig := &app.AppConfig{
		Type:           "http",
		Port:           os.Getenv("PORT"),
		AuthDBPath:     authDBPath,
		EncryptionKey:  encryptionKey,
		ZentaoServer:   zentaoServer,
		ZentaoAccount:  zentaoAccount,
		ZentaoPassword: zentaoPassword,
		StaticPath:     "embedded", // 标记使用嵌入的静态资源
	}

	// 使用依赖注入初始化应用
	application, err := app.InitializeHTTPApp(appConfig)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// 类型断言获取HTTPApp
	httpApp, ok := application.(*app.HTTPApp)
	if !ok {
		log.Fatalf("Expected HTTPApp, got %T", application)
	}

	// 启动应用（这会初始化路由）
	if err := httpApp.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// 获取路由器并添加静态资源处理
	router := httpApp.GetRouter()
	if router == nil {
		log.Fatal("Failed to get router")
	}

	// 挂载前端静态资源到根目录
	staticFS := getFileSystem()

	// 先注册API路由，再注册静态资源路由
	// 使用自定义的静态文件处理，避免通配符路由冲突
	router.Use(func(c *gin.Context) {
		// 检查是否是API请求
		if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:4] == "/api" {
			// 是API请求，继续处理
			c.Next()
			return
		}

		// 是静态资源请求，尝试从静态文件系统中获取
		filePath := c.Request.URL.Path
		if filePath == "/" {
			filePath = "index.html"
		} else {
			// 移除开头的斜杠
			filePath = filePath[1:]
		}

		// 尝试打开文件
		file, err := staticFS.Open(filePath)
		if err == nil {
			// 文件存在，返回文件
			defer file.Close()
			fileInfo, err := file.Stat()
			if err == nil && !fileInfo.IsDir() {
				// 根据文件类型设置Content-Type
				ext := filepath.Ext(filePath)
				contentType := mime.TypeByExtension(ext)
				if contentType == "" {
					contentType = "application/octet-stream"
				}
				c.Header("Content-Type", contentType)
				c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
				c.Status(http.StatusOK)
				io.Copy(c.Writer, file)
				c.Abort()
				return
			}
		}

		// 文件不存在，返回index.html（用于SPA路由）
		indexFile, err := staticFS.Open("index.html")
		if err == nil {
			defer indexFile.Close()
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Status(http.StatusOK)
			io.Copy(c.Writer, indexFile)
			c.Abort()
			return
		}

		// 如果index.html也不存在，继续处理
		c.Next()
	})

	log.Printf("Serving frontend from embedded static file system")

	// 等待应用退出
	select {}
}
