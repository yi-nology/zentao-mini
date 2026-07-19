package handlers

import (
	"log"

	"github.com/yi-nology/zentao-mini/backend/core/initialization"
	"github.com/yi-nology/zentao-mini/backend/core/service"
	myzentao "github.com/yi-nology/zentao-mini/backend/core/zentao"
)

// HandlerRegistry Handler注册表
// 使用单例模式确保所有handler和service只初始化一次，避免重复创建
// 这是依赖注入模式的核心组件，提供统一的handler访问入口
type HandlerRegistry struct {
	client *myzentao.Client

	productService   *service.ProductService
	projectService   *service.ProjectService
	executionService *service.ExecutionService
	bugService       *service.BugService
	buildService     *service.BuildService
	storyService     *service.StoryService
	taskService      *service.TaskService
	userService      *service.UserService
	timelogService   *service.TimelogService
	dashboardService *service.DashboardService
	reportService    *service.ReportService
	webhookService   *service.WebhookService
	schedulerService *service.SchedulerService

	productHandler   *ProductHandler
	projectHandler   *ProjectHandler
	executionHandler *ExecutionHandler
	bugHandler       *BugHandler
	buildHandler     *BuildHandler
	storyHandler     *StoryHandler
	taskHandler      *TaskHandler
	userHandler      *UserHandler
	timelogHandler   *TimelogHandler
	dashboardHandler *DashboardHandler
	schedulerHandler *SchedulerHandler
	logHandler       *LogHandler
	initHandler      *InitHandler
	healthHandler    *HealthHandler
}

// NewHandlerRegistry 创建Handler注册表
// 所有service和handler在此处统一初始化，确保整个应用生命周期内只创建一次
// 采用分层架构：Client -> Service -> Handler
func NewHandlerRegistry(client *myzentao.Client, initService *initialization.InitService) *HandlerRegistry {
	registry := &HandlerRegistry{
		client: client,
	}

	registry.productService = service.NewProductService(client)
	registry.projectService = service.NewProjectService(client)
	registry.executionService = service.NewExecutionService(client)
	registry.bugService = service.NewBugService(client)
	registry.buildService = service.NewBuildService(client)
	registry.storyService = service.NewStoryService(client)
	registry.taskService = service.NewTaskService(client)
	registry.userService = service.NewUserService(client)
	registry.timelogService = service.NewTimelogService(client)
	registry.dashboardService = service.NewDashboardService(client)

	registry.productHandler = NewProductHandler(registry.productService)
	registry.projectHandler = NewProjectHandler(registry.projectService)
	registry.executionHandler = NewExecutionHandler(registry.executionService)
	registry.bugHandler = NewBugHandler(registry.bugService)
	registry.buildHandler = NewBuildHandler(registry.buildService)
	registry.storyHandler = NewStoryHandler(registry.storyService)
	registry.taskHandler = NewTaskHandler(registry.taskService)
	registry.userHandler = NewUserHandler(registry.userService)
	registry.timelogHandler = NewTimelogHandler(registry.timelogService)
	registry.dashboardHandler = NewDashboardHandler(registry.dashboardService)

	registry.logHandler = NewLogHandler()

	registry.initHandler = NewInitHandler(initService, client)

	registry.healthHandler = NewHealthHandler(
		client,
		registry.productService,
		registry.projectService,
		registry.bugService,
		registry.storyService,
		registry.taskService,
		registry.userService,
		nil,
	)

	return registry
}

func (r *HandlerRegistry) InitScheduler(store *initialization.ConfigStore) {
	r.reportService = service.NewReportService(r.client)
	r.webhookService = service.NewWebhookService()
	r.schedulerService = service.NewSchedulerService(store, r.reportService, r.webhookService)
	r.schedulerHandler = NewSchedulerHandler(r.schedulerService, r.webhookService, r.reportService)
	r.healthHandler.SetSchedulerService(r.schedulerService)
	if err := r.schedulerService.Start(); err != nil {
		log.Printf("Failed to start scheduler: %v", err)
	}
}

func (r *HandlerRegistry) StopScheduler() {
	if r.schedulerService != nil {
		r.schedulerService.Stop()
	}
}

// GetProductHandler 获取产品Handler
func (r *HandlerRegistry) GetProductHandler() *ProductHandler {
	return r.productHandler
}

// GetProjectHandler 获取项目Handler
func (r *HandlerRegistry) GetProjectHandler() *ProjectHandler {
	return r.projectHandler
}

// GetExecutionHandler 获取执行Handler
func (r *HandlerRegistry) GetExecutionHandler() *ExecutionHandler {
	return r.executionHandler
}

// GetBugHandler 获取Bug Handler
func (r *HandlerRegistry) GetBugHandler() *BugHandler {
	return r.bugHandler
}

// GetBuildHandler 获取版本 Handler
func (r *HandlerRegistry) GetBuildHandler() *BuildHandler {
	return r.buildHandler
}

// GetStoryHandler 获取需求Handler
func (r *HandlerRegistry) GetStoryHandler() *StoryHandler {
	return r.storyHandler
}

// GetTaskHandler 获取任务Handler
func (r *HandlerRegistry) GetTaskHandler() *TaskHandler {
	return r.taskHandler
}

// GetUserHandler 获取用户Handler
func (r *HandlerRegistry) GetUserHandler() *UserHandler {
	return r.userHandler
}

// GetTimelogHandler 获取工时Handler
func (r *HandlerRegistry) GetTimelogHandler() *TimelogHandler {
	return r.timelogHandler
}

// GetProductService 获取产品 Service
func (r *HandlerRegistry) GetProductService() *service.ProductService { return r.productService }

// GetProjectService 获取项目 Service
func (r *HandlerRegistry) GetProjectService() *service.ProjectService { return r.projectService }

// GetExecutionService 获取执行 Service
func (r *HandlerRegistry) GetExecutionService() *service.ExecutionService { return r.executionService }

// GetBugService 获取Bug Service
func (r *HandlerRegistry) GetBugService() *service.BugService { return r.bugService }

// GetBuildService 获取版本 Service
func (r *HandlerRegistry) GetBuildService() *service.BuildService { return r.buildService }

// GetStoryService 获取需求 Service
func (r *HandlerRegistry) GetStoryService() *service.StoryService { return r.storyService }

// GetTaskService 获取任务 Service
func (r *HandlerRegistry) GetTaskService() *service.TaskService { return r.taskService }

// GetUserService 获取用户 Service
func (r *HandlerRegistry) GetUserService() *service.UserService { return r.userService }

// GetTimelogService 获取工时 Service
func (r *HandlerRegistry) GetTimelogService() *service.TimelogService { return r.timelogService }

// GetInitHandler 获取初始化Handler
func (r *HandlerRegistry) GetInitHandler() *InitHandler {
	return r.initHandler
}

// GetLogHandler 获取日志 Handler
func (r *HandlerRegistry) GetLogHandler() *LogHandler {
	return r.logHandler
}

func (r *HandlerRegistry) GetDashboardHandler() *DashboardHandler {
	return r.dashboardHandler
}

func (r *HandlerRegistry) GetSchedulerHandler() *SchedulerHandler {
	return r.schedulerHandler
}

func (r *HandlerRegistry) GetHealthHandler() *HealthHandler {
	return r.healthHandler
}
