package zentao

import (
	"github.com/yi-nology/zentao-mini/backend/core/handlers"
	coremcp "github.com/yi-nology/zentao-mini/backend/core/mcp"
	"github.com/yi-nology/zentao-mini/backend/core/version"
)

var (
	productHandler    *handlers.ProductHandler
	projectHandler    *handlers.ProjectHandler
	executionHandler  *handlers.ExecutionHandler
	bugHandler        *handlers.BugHandler
	storyHandler      *handlers.StoryHandler
	taskHandler       *handlers.TaskHandler
	userHandler       *handlers.UserHandler
	timelogHandler    *handlers.TimelogHandler
	dashboardHandler  *handlers.DashboardHandler
	schedulerHandler  *handlers.SchedulerHandler
	initHandler       *handlers.InitHandler
	healthHandler     *handlers.HealthHandler
	mcpHTTPTransport  *coremcp.HTTPTransport
)

func Init(registry *handlers.HandlerRegistry, transport *coremcp.HTTPTransport) {
	productHandler = registry.GetProductHandler()
	projectHandler = registry.GetProjectHandler()
	executionHandler = registry.GetExecutionHandler()
	bugHandler = registry.GetBugHandler()
	storyHandler = registry.GetStoryHandler()
	taskHandler = registry.GetTaskHandler()
	userHandler = registry.GetUserHandler()
	timelogHandler = registry.GetTimelogHandler()
	dashboardHandler = registry.GetDashboardHandler()
	schedulerHandler = registry.GetSchedulerHandler()
	initHandler = registry.GetInitHandler()
	healthHandler = registry.GetHealthHandler()
	mcpHTTPTransport = transport
}

var versionInfo = version.Info
