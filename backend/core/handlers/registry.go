package handlers

import (
	"chandao-mini/backend/core/service"
	myzentao "chandao-mini/backend/core/zentao"
)

// HandlerRegistry Handler注册表
// 使用单例模式确保所有handler和service只初始化一次，避免重复创建
// 这是依赖注入模式的核心组件，提供统一的handler访问入口
type HandlerRegistry struct {
	// 禅道客户端
	client *myzentao.Client

	// Service层实例
	productService   *service.ProductService
	projectService   *service.ProjectService
	executionService *service.ExecutionService
	bugService       *service.BugService
	storyService     *service.StoryService
	taskService      *service.TaskService
	userService      *service.UserService
	timelogService   *service.TimelogService

	// Handler层实例
	productHandler   *ProductHandler
	projectHandler   *ProjectHandler
	executionHandler *ExecutionHandler
	bugHandler       *BugHandler
	storyHandler     *StoryHandler
	taskHandler      *TaskHandler
	userHandler      *UserHandler
	timelogHandler   *TimelogHandler
	mcpHandler       *MCPHandler
}

// NewHandlerRegistry 创建Handler注册表
// 所有service和handler在此处统一初始化，确保整个应用生命周期内只创建一次
// 采用分层架构：Client -> Service -> Handler
func NewHandlerRegistry(client *myzentao.Client) *HandlerRegistry {
	registry := &HandlerRegistry{
		client: client,
	}

	// 初始化所有Service层实例
	registry.productService = service.NewProductService(client)
	registry.projectService = service.NewProjectService(client)
	registry.executionService = service.NewExecutionService(client)
	registry.bugService = service.NewBugService(client)
	registry.storyService = service.NewStoryService(client)
	registry.taskService = service.NewTaskService(client)
	registry.userService = service.NewUserService(client)
	registry.timelogService = service.NewTimelogService(client)

	// 初始化所有Handler层实例，注入Service依赖
	registry.productHandler = NewProductHandler(registry.productService)
	registry.projectHandler = NewProjectHandler(registry.projectService)
	registry.executionHandler = NewExecutionHandler(registry.executionService)
	registry.bugHandler = NewBugHandler(registry.bugService)
	registry.storyHandler = NewStoryHandler(registry.storyService)
	registry.taskHandler = NewTaskHandler(registry.taskService)
	registry.userHandler = NewUserHandler(registry.userService)
	registry.timelogHandler = NewTimelogHandler(registry.timelogService)

	// MCP handler依赖其他handler
	registry.mcpHandler = NewMCPHandler(
		registry.productHandler,
		registry.projectHandler,
		registry.executionHandler,
		registry.bugHandler,
		registry.storyHandler,
		registry.taskHandler,
		registry.userHandler,
		registry.timelogHandler,
	)

	return registry
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

// GetMCPHandler 获取MCP Handler
func (r *HandlerRegistry) GetMCPHandler() *MCPHandler {
	return r.mcpHandler
}
