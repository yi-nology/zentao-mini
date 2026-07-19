package routes

import (
	"context"
	"net/http"
	"path"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/route"

	bizhandler "github.com/yi-nology/zentao-mini/backend/biz/handler/zentao"
	bizrouter "github.com/yi-nology/zentao-mini/backend/biz/router"
	"github.com/yi-nology/zentao-mini/backend/core/config"
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
	return SetupRouterWithHandlers(initService, zentaoClient, registry, ":12345", nil)
}

func SetupRouterWithHandlers(initService *initialization.InitService, zentaoClient *zentao.Client, registry *handlers.HandlerRegistry, hostPort string, staticFS http.FileSystem) *server.Hertz {
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

	// 初始化 MCP 模式管理器：从配置加载，HTTP 入口标记传输模式为 http
	mcpCfg := config.Get().MCP
	mcp.GetMCPModeManager().InitFromConfig(mcpCfg)
	mcp.GetMCPModeManager().SetTransport(mcp.TransportHTTP)

	// both 模式：HTTP 入口额外启动后台 stdio 监听，与 HTTP 共享同一 MCPServer
	if mcpCfg.Transport == mcp.TransportBoth {
		mcp.NewStdioTransport(mcpServer).Start()
		logger.Info("MCP stdio transport started in 'both' mode")
	}

	bizhandler.Init(registry, mcpTransport)

	bizrouter.GeneratedRegister(hertzServer)

	registerCustomRoutes(hertzServer, registry, mcpTransport)

	if staticFS != nil {
		hertzServer.NoRoute(func(ctx context.Context, c *app.RequestContext) {
			filePath := string(c.Path())
			if filePath == "/" {
				filePath = "/index.html"
			}

			f, err := staticFS.Open(filePath)
			if err != nil {
				// SPA fallback: serve index.html for non-file routes
				f, err = staticFS.Open("/index.html")
				if err != nil {
					c.SetStatusCode(404)
					return
				}
			}
			defer func() { _ = f.Close() }()

			stat, err := f.Stat()
			if err != nil || stat.IsDir() {
				// For directories, try index.html
				_ = f.Close()
				f, err = staticFS.Open("/index.html")
				if err != nil {
					c.SetStatusCode(404)
					return
				}
				defer func() { _ = f.Close() }()
				stat, _ = f.Stat()
			}

			data := make([]byte, stat.Size())
			_, _ = f.Read(data)

			ext := getExt(filePath)
			if ct := getContentType(ext); ct != "" {
				c.Header("Content-Type", ct)
			}
			c.Header("Cache-Control", "public, max-age=3600")
			c.SetStatusCode(200)
			_, _ = c.Write(data)
		})
		logger.Info("Static file system mounted")
	}

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
	registerMCPAdminRoutes(r)
}

// registerMCPAdminRoutes 注册 MCP 运行时管理 API（热重载 / 状态查询）
func registerMCPAdminRoutes(r *server.Hertz) {
	admin := mcp.NewMCPAdminHandler()
	api := r.Group("/api/v1")
	admin.RegisterAdminRoutes(api)
}

func registerBackwardCompatRoutes(r *server.Hertz, registry *handlers.HandlerRegistry) {
	api := r.Group("/api")

	api.POST("/init/upload", registry.GetInitHandler().UploadConfig)
	api.GET("/init/status", bizhandler.GetInitStatus)
	api.GET("/init/account", bizhandler.GetAccountInfo)

	registerDomainRoutes(api)
	registerSchedulerRoutes(api, registry.GetSchedulerHandler())

	// 日志查看接口（供前端日志页消费）
	logHandler := registry.GetLogHandler()
	api.GET("/logs", logHandler.GetLogs)
	api.DELETE("/logs", logHandler.ClearLogs)
	api.GET("/logs/status", logHandler.LogsStatus)

	// 离线缓存管理接口（缓存初始化失败时跳过注册）
	if cacheH := registry.GetCacheHandler(); cacheH != nil {
		api.GET("/cache/status", cacheH.GetStatus)
		api.DELETE("/cache", cacheH.ClearAll)
		api.DELETE("/cache/:entityType", cacheH.Invalidate)
	}
}

func registerDomainRoutes(g *route.RouterGroup) {
	g.GET("/products", bizhandler.GetProducts)
	g.GET("/projects", bizhandler.GetProjects)
	g.GET("/executions", bizhandler.GetExecutions)
	g.GET("/bugs", bizhandler.GetBugs)
	g.GET("/builds/project", bizhandler.GetBuildsByProject)
	g.GET("/builds/execution", bizhandler.GetBuildsByExecution)
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

func getExt(filePath string) string {
	return strings.ToLower(path.Ext(filePath))
}

func getContentType(ext string) string {
	types := map[string]string{
		".html":  "text/html; charset=utf-8",
		".css":   "text/css; charset=utf-8",
		".js":    "application/javascript; charset=utf-8",
		".json":  "application/json",
		".png":   "image/png",
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".gif":   "image/gif",
		".svg":   "image/svg+xml",
		".ico":   "image/x-icon",
		".woff":  "font/woff",
		".woff2": "font/woff2",
		".ttf":   "font/ttf",
		".eot":   "application/vnd.ms-fontobject",
		".map":   "application/json",
	}
	if ct, ok := types[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}
