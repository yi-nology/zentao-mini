package mcp

import (
	"chandao-mini/backend/core/service"
	myzentao "chandao-mini/backend/core/zentao"
)

// MCPServer 是 MCP 服务的核心，直接依赖 service 层
// 统一处理所有 action，消除 stdio/HTTP 的代码重复
type MCPServer struct {
	productService   *service.ProductService
	projectService   *service.ProjectService
	executionService *service.ExecutionService
	bugService       *service.BugService
	storyService     *service.StoryService
	taskService      *service.TaskService
	userService      *service.UserService
	timelogService   *service.TimelogService
}

// NewMCPServer 创建 MCP 服务实例
func NewMCPServer(client *myzentao.Client) *MCPServer {
	return &MCPServer{
		productService:   service.NewProductService(client),
		projectService:   service.NewProjectService(client),
		executionService: service.NewExecutionService(client),
		bugService:       service.NewBugService(client),
		storyService:     service.NewStoryService(client),
		taskService:      service.NewTaskService(client),
		userService:      service.NewUserService(client),
		timelogService:   service.NewTimelogService(client),
	}
}

// NewMCPServerFromServices 从已有的 service 实例创建 MCP 服务
func NewMCPServerFromServices(
	productService *service.ProductService,
	projectService *service.ProjectService,
	executionService *service.ExecutionService,
	bugService *service.BugService,
	storyService *service.StoryService,
	taskService *service.TaskService,
	userService *service.UserService,
	timelogService *service.TimelogService,
) *MCPServer {
	return &MCPServer{
		productService:   productService,
		projectService:   projectService,
		executionService: executionService,
		bugService:       bugService,
		storyService:     storyService,
		taskService:      taskService,
		userService:      userService,
		timelogService:   timelogService,
	}
}

// HandleAction 统一入口，处理所有 MCP action
func (s *MCPServer) HandleAction(action string, params map[string]interface{}) (interface{}, error) {
	switch action {
	case "ping":
		return s.handlePing(params)
	case "get_products":
		return s.handleGetProducts(params)
	case "get_projects":
		return s.handleGetProjects(params)
	case "get_executions":
		return s.handleGetExecutions(params)
	case "get_bugs":
		return s.handleGetBugs(params)
	case "get_stories":
		return s.handleGetStories(params)
	case "get_tasks":
		return s.handleGetTasks(params)
	case "get_users":
		return s.handleGetUsers(params)
	case "get_timelog":
		return s.handleGetTimelog(params)
	default:
		return nil, &ActionError{Action: action, Message: "unknown action"}
	}
}

// ActionError 未知 action 错误
type ActionError struct {
	Action  string
	Message string
}

func (e *ActionError) Error() string {
	return e.Message + ": " + e.Action
}
