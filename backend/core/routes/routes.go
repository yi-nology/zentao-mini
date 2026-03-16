package routes

import (
	"os"

	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/handlers"
	"chandao-mini/backend/core/initialization"
	"chandao-mini/backend/core/logger"
	"chandao-mini/backend/core/metrics"
	"chandao-mini/backend/core/middleware"
	"chandao-mini/backend/core/utils"
	"chandao-mini/backend/core/zentao"

	"go.uber.org/zap"
)

// SetupRouter configures the router with all routes
// 已废弃：请使用 SetupRouterWithHandlers 以避免重复创建handler
func SetupRouter(initService *initialization.InitService, zentaoClient *zentao.Client) *gin.Engine {
	registry := handlers.NewHandlerRegistry(zentaoClient, initService)
	return SetupRouterWithHandlers(initService, zentaoClient, registry)
}

// SetupRouterWithHandlers 使用HandlerRegistry配置路由
// 推荐使用此方法，确保handler单例模式
func SetupRouterWithHandlers(initService *initialization.InitService, zentaoClient *zentao.Client, registry *handlers.HandlerRegistry) *gin.Engine {
	// 设置 Gin 模式
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	// 创建 Gin 引擎
	r := gin.New() // 使用 gin.New() 而不是 gin.Default() 以便自定义中间件

	// 添加全局中间件（按顺序执行）
	r.Use(middleware.RecoveryMiddleware()) // Panic恢复中间件（必须放在最前面）
	r.Use(middleware.TraceIDMiddleware())  // 请求追踪ID中间件
	r.Use(middleware.LoggerMiddleware())   // 日志中间件
	r.Use(middleware.MetricsMiddleware())  // 性能监控中间件
	r.Use(errors.RateLimitMiddleware())    // 请求限流中间件
	r.Use(utils.PaginationMiddleware())    // 分页中间件
	r.Use(errors.CORSMiddleware())         // CORS中间件（从环境变量读取配置）

	// 从注册表获取处理器（单例模式）
	initHandler := registry.GetInitHandler()
	productHandler := registry.GetProductHandler()
	projectHandler := registry.GetProjectHandler()
	executionHandler := registry.GetExecutionHandler()
	bugHandler := registry.GetBugHandler()
	storyHandler := registry.GetStoryHandler()
	taskHandler := registry.GetTaskHandler()
	userHandler := registry.GetUserHandler()
	timelogHandler := registry.GetTimelogHandler()

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		errors.Success(c, gin.H{
			"status":  "ok",
			"message": "chandao-mini backend is running",
		})
	})

	// Prometheus metrics端点
	r.GET("/metrics", metrics.Handler())

	// 注册API路由（支持版本控制和向后兼容）
	registerAPIRoutes(r, initHandler, productHandler, projectHandler, executionHandler, bugHandler, storyHandler, taskHandler, userHandler, timelogHandler)

	logger.Info("Router setup completed", zap.String("gin_mode", ginMode))

	return r
}

// registerAPIRoutes 注册API路由
// 同时支持 /api/v1 和 /api 路由，保持向后兼容
func registerAPIRoutes(
	r *gin.Engine,
	initHandler *handlers.InitHandler,
	productHandler *handlers.ProductHandler,
	projectHandler *handlers.ProjectHandler,
	executionHandler *handlers.ExecutionHandler,
	bugHandler *handlers.BugHandler,
	storyHandler *handlers.StoryHandler,
	taskHandler *handlers.TaskHandler,
	userHandler *handlers.UserHandler,
	timelogHandler *handlers.TimelogHandler,
) {
	registerRoutes := func(apiGroup *gin.RouterGroup) {
		// 初始化相关接口
		apiGroup.POST("/init/upload", initHandler.UploadConfig)
		apiGroup.GET("/init/status", initHandler.GetInitStatus)

		// 产品相关接口
		apiGroup.GET("/products", productHandler.GetProducts)

		// 项目相关接口
		apiGroup.GET("/projects", projectHandler.GetProjects)

		// 执行/迭代相关接口
		apiGroup.GET("/executions", executionHandler.GetExecutions)

		// Bug相关接口
		apiGroup.GET("/bugs", bugHandler.GetBugs)

		// 需求相关接口
		apiGroup.GET("/stories", storyHandler.GetStories)

		// 任务相关接口
		apiGroup.GET("/tasks", taskHandler.GetTasks)

		// 用户相关接口
		apiGroup.GET("/users", userHandler.GetUsers)
		apiGroup.GET("/users/all", userHandler.GetUsersAll)
		apiGroup.GET("/users/current", userHandler.GetCurrentUser)

		// 工时统计相关接口
		apiGroup.GET("/timelog/analysis", timelogHandler.GetTimelogAnalysis)
		apiGroup.GET("/timelog/dashboard", timelogHandler.GetTimelogDashboard)
		apiGroup.GET("/timelog/efforts", timelogHandler.GetTimelogEfforts)
	}

	// 注册API v1版本路由（推荐使用）
	v1 := r.Group("/api/v1")
	registerRoutes(v1)

	// 注册API路由（向后兼容，保持原有路由）
	// 注意：新客户端应使用 /api/v1 路由
	api := r.Group("/api")
	registerRoutes(api)
}
