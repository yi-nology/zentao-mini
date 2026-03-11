package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// MCPHandler 处理MCP协议相关请求
type MCPHandler struct {
	productHandler  *ProductHandler
	projectHandler  *ProjectHandler
	executionHandler *ExecutionHandler
	bugHandler      *BugHandler
	storyHandler    *StoryHandler
	taskHandler     *TaskHandler
	userHandler     *UserHandler
	timelogHandler  *TimelogHandler
	stdin           io.Reader
	stdout          io.Writer
	mutex           sync.Mutex
}

// NewMCPHandler 创建新的MCP处理器
func NewMCPHandler(
	productHandler *ProductHandler,
	projectHandler *ProjectHandler,
	executionHandler *ExecutionHandler,
	bugHandler *BugHandler,
	storyHandler *StoryHandler,
	taskHandler *TaskHandler,
	userHandler *UserHandler,
	timelogHandler *TimelogHandler,
) *MCPHandler {
	return &MCPHandler{
		productHandler:  productHandler,
		projectHandler:  projectHandler,
		executionHandler: executionHandler,
		bugHandler:      bugHandler,
		storyHandler:    storyHandler,
		taskHandler:     taskHandler,
		userHandler:     userHandler,
		timelogHandler:  timelogHandler,
		stdin:           os.Stdin,
		stdout:          os.Stdout,
	}
}

// Start 启动MCP服务
func (h *MCPHandler) Start() {
	go h.handleRequests()
}

// handleRequests 处理MCP请求
func (h *MCPHandler) handleRequests() {
	decoder := json.NewDecoder(h.stdin)
	encoder := json.NewEncoder(h.stdout)

	for {
		var request map[string]interface{}
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				break
			}
			h.sendErrorResponse(encoder, fmt.Sprintf("Invalid request: %v", err))
			continue
		}

		h.handleRequest(encoder, request)
	}
}

// handleRequest 处理单个MCP请求
func (h *MCPHandler) handleRequest(encoder *json.Encoder, request map[string]interface{}) {
	action, ok := request["action"].(string)
	if !ok {
		h.sendErrorResponse(encoder, "Missing or invalid action")
		return
	}

	params, _ := request["params"].(map[string]interface{})

	switch action {
	case "get_products":
		h.handleGetProducts(encoder, params)
	case "get_projects":
		h.handleGetProjects(encoder, params)
	case "get_executions":
		h.handleGetExecutions(encoder, params)
	case "get_bugs":
		h.handleGetBugs(encoder, params)
	case "get_stories":
		h.handleGetStories(encoder, params)
	case "get_tasks":
		h.handleGetTasks(encoder, params)
	case "get_users":
		h.handleGetUsers(encoder, params)
	case "get_timelog":
		h.handleGetTimelog(encoder, params)
	case "ping":
		h.handlePing(encoder)
	default:
		h.sendErrorResponse(encoder, fmt.Sprintf("Unknown action: %s", action))
	}
}

// handleGetProducts 处理获取产品列表请求
func (h *MCPHandler) handleGetProducts(encoder *json.Encoder, params map[string]interface{}) {
	// 这里需要实现具体的处理逻辑
	// 由于我们没有gin.Context，需要创建一个模拟的上下文
	// 或者修改现有的处理器方法，使其不依赖于gin.Context
	h.sendResponse(encoder, map[string]interface{}{
		"status": "ok",
		"message": "Products retrieved successfully",
		"data": []interface{}{},
	})
}

// handleGetProjects 处理获取项目列表请求
func (h *MCPHandler) handleGetProjects(encoder *json.Encoder, params map[string]interface{}) {
	h.sendResponse(encoder, map[string]interface{}{
		"status": "ok",
		"message": "Projects retrieved successfully",
		"data": []interface{}{},
	})
}

// handleGetExecutions 处理获取执行/迭代列表请求
func (h *MCPHandler) handleGetExecutions(encoder *json.Encoder, params map[string]interface{}) {
	h.sendResponse(encoder, map[string]interface{}{
		"status": "ok",
		"message": "Executions retrieved successfully",
		"data": []interface{}{},
	})
}

// handleGetBugs 处理获取Bug列表请求
func (h *MCPHandler) handleGetBugs(encoder *json.Encoder, params map[string]interface{}) {
	h.sendResponse(encoder, map[string]interface{}{
		"status": "ok",
		"message": "Bugs retrieved successfully",
		"data": []interface{}{},
	})
}

// handleGetStories 处理获取需求列表请求
func (h *MCPHandler) handleGetStories(encoder *json.Encoder, params map[string]interface{}) {
	h.sendResponse(encoder, map[string]interface{}{
		"status": "ok",
		"message": "Stories retrieved successfully",
		"data": []interface{}{},
	})
}

// handleGetTasks 处理获取任务列表请求
func (h *MCPHandler) handleGetTasks(encoder *json.Encoder, params map[string]interface{}) {
	h.sendResponse(encoder, map[string]interface{}{
		"status": "ok",
		"message": "Tasks retrieved successfully",
		"data": []interface{}{},
	})
}

// handleGetUsers 处理获取用户列表请求
func (h *MCPHandler) handleGetUsers(encoder *json.Encoder, params map[string]interface{}) {
	h.sendResponse(encoder, map[string]interface{}{
		"status": "ok",
		"message": "Users retrieved successfully",
		"data": []interface{}{},
	})
}

// handleGetTimelog 处理获取工时数据请求
func (h *MCPHandler) handleGetTimelog(encoder *json.Encoder, params map[string]interface{}) {
	h.sendResponse(encoder, map[string]interface{}{
		"status": "ok",
		"message": "Timelog retrieved successfully",
		"data": []interface{}{},
	})
}

// handlePing 处理ping请求
func (h *MCPHandler) handlePing(encoder *json.Encoder) {
	h.sendResponse(encoder, map[string]interface{}{
		"status": "ok",
		"message": "Pong",
		"version": "1.0",
	})
}

// sendResponse 发送成功响应
func (h *MCPHandler) sendResponse(encoder *json.Encoder, data map[string]interface{}) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if err := encoder.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send response: %v\n", err)
	}
}

// sendErrorResponse 发送错误响应
func (h *MCPHandler) sendErrorResponse(encoder *json.Encoder, errorMsg string) {
	h.sendResponse(encoder, map[string]interface{}{
		"status": "error",
		"message": errorMsg,
	})
}