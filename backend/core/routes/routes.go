package routes

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/route"

	bizhandler "github.com/yi-nology/zentao-mini/backend/biz/handler/zentao"
	bizrouter "github.com/yi-nology/zentao-mini/backend/biz/router"
	"github.com/yi-nology/zentao-mini/backend/core/errors"
	"github.com/yi-nology/zentao-mini/backend/core/handlers"
	"github.com/yi-nology/zentao-mini/backend/core/initialization"
	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/mcp"
	"github.com/yi-nology/zentao-mini/backend/core/metrics"
	"github.com/yi-nology/zentao-mini/backend/core/middleware"
	"github.com/yi-nology/zentao-mini/backend/core/utils"
	"github.com/yi-nology/zentao-mini/backend/core/zentao"
)

func SetupRouter(initService *initialization.InitService, zentaoClient *zentao.Client) *server.Hertz {
	registry := handlers.NewHandlerRegistry(zentaoClient, initService)
	return SetupRouterWithHandlers(initService, zentaoClient, registry, ":12345")
}

func SetupRouterWithHandlers(initService *initialization.InitService, zentaoClient *zentao.Client, registry *handlers.HandlerRegistry, hostPort string) *server.Hertz {
	hertzServer := server.New(server.WithHostPorts(hostPort))

	hertzServer.Use(middleware.RecoveryMiddleware())
	hertzServer.Use(middleware.TraceIDMiddleware())
	hertzServer.Use(middleware.LoggerMiddleware())
	hertzServer.Use(middleware.MetricsMiddleware())
	hertzServer.Use(errors.RateLimitMiddleware())
	hertzServer.Use(utils.PaginationMiddleware())
	hertzServer.Use(errors.CORSMiddleware())

	mcpServer := mcp.NewMCPServerFromServices(
		registry.GetProductService(),
		registry.GetProjectService(),
		registry.GetExecutionService(),
		registry.GetBugService(),
		registry.GetStoryService(),
		registry.GetTaskService(),
		registry.GetUserService(),
		registry.GetTimelogService(),
	)
	mcpTransport := mcp.NewHTTPTransport(mcpServer)

	bizhandler.Init(registry, mcpTransport)

	bizrouter.GeneratedRegister(hertzServer)

	registerCustomRoutes(hertzServer, registry, mcpTransport)

	logger.Info("Router setup completed")

	return hertzServer
}

func registerCustomRoutes(r *server.Hertz, registry *handlers.HandlerRegistry, transport *mcp.HTTPTransport) {
	r.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		errors.Success(c, map[string]interface{}{
			"status":  "ok",
			"message": "zentao-mini backend is running",
		})
	})

	r.GET("/metrics", metrics.Handler())

	r.POST("/api/v1/init/upload", registry.GetInitHandler().UploadConfig)

	registerBackwardCompatRoutes(r, registry)
	registerMCPPostRoutes(r, transport)
}

func registerBackwardCompatRoutes(r *server.Hertz, registry *handlers.HandlerRegistry) {
	api := r.Group("/api")

	api.POST("/init/upload", registry.GetInitHandler().UploadConfig)
	api.GET("/init/status", bizhandler.GetInitStatus)
	api.GET("/init/account", bizhandler.GetAccountInfo)

	registerDomainRoutes(api)
	registerSchedulerRoutes(api, registry.GetSchedulerHandler())
}

func registerDomainRoutes(g *route.RouterGroup) {
	g.GET("/products", bizhandler.GetProducts)
	g.GET("/projects", bizhandler.GetProjects)
	g.GET("/executions", bizhandler.GetExecutions)
	g.GET("/bugs", bizhandler.GetBugs)
	g.GET("/stories", bizhandler.GetStories)
	g.GET("/tasks", bizhandler.GetTasks)
	g.GET("/users", bizhandler.GetUsers)
	g.GET("/users/all", bizhandler.GetUsersAll)
	g.GET("/users/current", bizhandler.GetCurrentUser)
	g.GET("/timelog/analysis", bizhandler.GetTimelogAnalysis)
	g.GET("/timelog/dashboard", bizhandler.GetTimelogDashboard)
	g.GET("/timelog/efforts", bizhandler.GetTimelogEfforts)
	g.GET("/dashboard", bizhandler.GetDashboard)
	g.GET("/project/overview", bizhandler.GetProjectOverview)
	g.GET("/personal/timelog", bizhandler.GetPersonalTimelog)
	g.GET("/search", bizhandler.Search)
}

func registerSchedulerRoutes(g *route.RouterGroup, h *handlers.SchedulerHandler) {
	scheduler := g.Group("/scheduler")
	scheduler.GET("/tasks", h.ListTasks)
	scheduler.POST("/tasks", h.CreateTask)
	scheduler.PUT("/tasks/:id", h.UpdateTask)
	scheduler.DELETE("/tasks/:id", h.DeleteTask)
	scheduler.PATCH("/tasks/:id/toggle", h.ToggleTask)
	scheduler.POST("/tasks/:id/run", h.RunTaskNow)
	scheduler.GET("/tasks/:id/logs", h.GetTaskLogs)
	scheduler.GET("/logs", h.GetAllLogs)
	scheduler.POST("/test-webhook", h.TestWebhook)
	scheduler.POST("/preview", h.PreviewReport)
}

func registerMCPPostRoutes(r *server.Hertz, transport *mcp.HTTPTransport) {
	actionMap := map[string]string{
		"ping":       "ping",
		"products":   "get_products",
		"projects":   "get_projects",
		"executions": "get_executions",
		"bugs":       "get_bugs",
		"stories":    "get_stories",
		"tasks":      "get_tasks",
		"users":      "get_users",
		"timelog":    "get_timelog",
	}

	for path, action := range actionMap {
		act := action
		r.POST("/mcp/"+path, func(ctx context.Context, c *app.RequestContext) {
			transport.HandleActionByName(ctx, act, c)
		})
	}
}
