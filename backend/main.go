package main

import (
	"log"
	"os"

	"chandao-mini/backend/core/handlers"
	"chandao-mini/backend/core/initialization"
	"chandao-mini/backend/core/routes"
	"chandao-mini/backend/core/zentao"
)

func main() {
	// 初始化服务
	initService := initialization.NewInitService(
		os.Getenv("AUTH_CONFIG_PATH"),
		os.Getenv("AUTH_DB_PATH"),
		os.Getenv("ENCRYPTION_KEY"),
	)

	// 加载禅道配置
	zentaoServer, zentaoAccount, zentaoPassword := initialization.LoadZentaoConfig(initService)

	zentaoClient := zentao.NewClient(zentaoServer, zentaoAccount, zentaoPassword)

	// 初始化处理器
	productHandler := handlers.NewProductHandler(zentaoClient)
	projectHandler := handlers.NewProjectHandler(zentaoClient)
	executionHandler := handlers.NewExecutionHandler(zentaoClient)
	bugHandler := handlers.NewBugHandler(zentaoClient)
	storyHandler := handlers.NewStoryHandler(zentaoClient)
	taskHandler := handlers.NewTaskHandler(zentaoClient)
	userHandler := handlers.NewUserHandler(zentaoClient)
	timelogHandler := handlers.NewTimelogHandler(zentaoClient)

	// 初始化MCP处理器（使用stdio通信）
	mcpHandler := handlers.NewMCPHandler(
		productHandler,
		projectHandler,
		executionHandler,
		bugHandler,
		storyHandler,
		taskHandler,
		userHandler,
		timelogHandler,
	)

	// 启动MCP服务
	mcpHandler.Start()

	// 设置路由
	r := routes.SetupRouter(initService, zentaoClient)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "12345"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Zentao Server: %s", zentaoServer)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
