package handlers

// Deprecated: 此文件中的 MCPHandler 已被 backend/core/mcp/ 包替代
// 新代码请使用 mcp.MCPServer + mcp.HTTPTransport / mcp.StdioTransport
// 保留此文件仅为向后兼容

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"chandao-mini/backend/core/errors"

	"github.com/gin-gonic/gin"
)

// MCPHandler 处理MCP协议相关请求
// 支持两种模式：stdio（标准输入输出）和 HTTP
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

// Start 启动MCP服务（stdio模式）
func (h *MCPHandler) Start() {
	go h.handleStdioRequests()
}

// handleStdioRequests 处理MCP stdio请求
func (h *MCPHandler) handleStdioRequests() {
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

// handleRequest 处理单个MCP请求（stdio模式）
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

// ============================================
// HTTP 模式处理方法
// ============================================

// MCPRequest HTTP MCP 请求结构
type MCPRequest struct {
	Action string                 `json:"action" binding:"required"`
	Params map[string]interface{} `json:"params"`
}

// MCPResponse HTTP MCP 响应结构
type MCPResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Version string      `json:"version,omitempty"`
}

// HandleMCPAction 处理MCP HTTP请求（统一入口）
func (h *MCPHandler) HandleMCPAction(c *gin.Context) {
	var req MCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	resp := h.processAction(req.Action, req.Params)
	if resp.Status == "error" {
		errors.ErrorWithCode(c, errors.CodeInternalError, resp.Message)
		} else {
		errors.Success(c, resp.Data)
	}
}

// HandleMCPActionGet 处理MCP HTTP GET请求（用于简单查询）
func (h *MCPHandler) HandleMCPActionGet(c *gin.Context) {
	action := c.Query("action")
	if action == "" {
		errors.BadRequest(c, "Missing 'action' query parameter")
		return
	}

	// 从查询参数构建 params
	params := make(map[string]interface{})
	for _, key := range []string{"productId", "projectId", "executionId", "status", "assignedTo", "dateFrom", "dateTo", "page", "pageSize"} {
		if val := c.Query(key); val != "" {
			params[key] = val
		}
	}

	resp := h.processAction(action, params)
	if resp.Status == "error" {
		errors.ErrorWithCode(c, errors.CodeInternalError, resp.Message)
	} else {
		errors.Success(c, resp.Data)
	}
}

// processAction 处理MCP action（HTTP模式共用）
func (h *MCPHandler) processAction(action string, params map[string]interface{}) MCPResponse {
	switch action {
	case "ping":
		return MCPResponse{
			Status:  "ok",
			Message: "Pong",
			Version: "1.0",
		}
	case "get_products":
		return h.processGetProducts(params)
	case "get_projects":
		return h.processGetProjects(params)
	case "get_executions":
		return h.processGetExecutions(params)
	case "get_bugs":
		return h.processGetBugs(params)
	case "get_stories":
		return h.processGetStories(params)
	case "get_tasks":
		return h.processGetTasks(params)
	case "get_users":
		return h.processGetUsers(params)
	case "get_timelog":
		return h.processGetTimelog(params)
	default:
		return MCPResponse{
			Status:  "error",
			Message: fmt.Sprintf("Unknown action: %s", action),
		}
	}
}

// ============================================
// 具体 action 处理方法
// ============================================

func (h *MCPHandler) processGetProducts(params map[string]interface{}) MCPResponse {
	result, err := h.productHandler.GetProductsHTTP()
	if err != nil {
		return MCPResponse{Status: "error", Message: err.Error()}
	}
	return MCPResponse{Status: "ok", Message: "Products retrieved successfully", Data: result}
}

func (h *MCPHandler) processGetProjects(params map[string]interface{}) MCPResponse {
	productId := ""
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	result, err := h.projectHandler.GetProjectsHTTP(productId)
	if err != nil {
		return MCPResponse{Status: "error", Message: err.Error()}
	}
	return MCPResponse{Status: "ok", Message: "Projects retrieved successfully", Data: result}
}

func (h *MCPHandler) processGetExecutions(params map[string]interface{}) MCPResponse {
	projectId := ""
	productId := ""
	if v, ok := params["projectId"]; ok {
		projectId = fmt.Sprintf("%v", v)
	}
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	result, err := h.executionHandler.GetExecutionsHTTP(projectId, productId)
	if err != nil {
		return MCPResponse{Status: "error", Message: err.Error()}
	}
	return MCPResponse{Status: "ok", Message: "Executions retrieved successfully", Data: result}
}

func (h *MCPHandler) processGetBugs(params map[string]interface{}) MCPResponse {
	productId := ""
	status := ""
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	if v, ok := params["status"]; ok {
		status = fmt.Sprintf("%v", v)
	}
	result, err := h.bugHandler.GetBugsHTTP(productId, status)
	if err != nil {
		return MCPResponse{Status: "error", Message: err.Error()}
	}
	return MCPResponse{Status: "ok", Message: "Bugs retrieved successfully", Data: result}
}

func (h *MCPHandler) processGetStories(params map[string]interface{}) MCPResponse {
	productId := ""
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	result, err := h.storyHandler.GetStoriesHTTP(productId)
	if err != nil {
		return MCPResponse{Status: "error", Message: err.Error()}
	}
	return MCPResponse{Status: "ok", Message: "Stories retrieved successfully", Data: result}
}

func (h *MCPHandler) processGetTasks(params map[string]interface{}) MCPResponse {
	productId := ""
	executionId := ""
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	if v, ok := params["executionId"]; ok {
		executionId = fmt.Sprintf("%v", v)
	}
	result, err := h.taskHandler.GetTasksHTTP(productId, executionId)
	if err != nil {
		return MCPResponse{Status: "error", Message: err.Error()}
	}
	return MCPResponse{Status: "ok", Message: "Tasks retrieved successfully", Data: result}
}

func (h *MCPHandler) processGetUsers(params map[string]interface{}) MCPResponse {
	result, err := h.userHandler.GetUsersHTTP()
	if err != nil {
		return MCPResponse{Status: "error", Message: err.Error()}
	}
	return MCPResponse{Status: "ok", Message: "Users retrieved successfully", Data: result}
}

func (h *MCPHandler) processGetTimelog(params map[string]interface{}) MCPResponse {
	productId := ""
	dateFrom := ""
	dateTo := ""
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	if v, ok := params["dateFrom"]; ok {
		dateFrom = fmt.Sprintf("%v", v)
	}
	if v, ok := params["dateTo"]; ok {
		dateTo = fmt.Sprintf("%v", v)
	}
	result, err := h.timelogHandler.GetTimelogHTTP(productId, dateFrom, dateTo)
	if err != nil {
		return MCPResponse{Status: "error", Message: err.Error()}
	}
	return MCPResponse{Status: "ok", Message: "Timelog retrieved successfully", Data: result}
}

// ============================================
// stdio 模式具体处理方法
// ============================================

func (h *MCPHandler) handleGetProducts(encoder *json.Encoder, params map[string]interface{}) {
	result, err := h.productHandler.GetProductsHTTP()
	if err != nil {
		h.sendErrorResponse(encoder, err.Error())
		return
	}
	h.sendResponse(encoder, map[string]interface{}{
		"status":  "ok",
		"message": "Products retrieved successfully",
		"data":    result,
	})
}

func (h *MCPHandler) handleGetProjects(encoder *json.Encoder, params map[string]interface{}) {
	productId := ""
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	result, err := h.projectHandler.GetProjectsHTTP(productId)
	if err != nil {
		h.sendErrorResponse(encoder, err.Error())
		return
	}
	h.sendResponse(encoder, map[string]interface{}{
		"status":  "ok",
		"message": "Projects retrieved successfully",
		"data":    result,
	})
}

func (h *MCPHandler) handleGetExecutions(encoder *json.Encoder, params map[string]interface{}) {
	projectId := ""
	productId := ""
	if v, ok := params["projectId"]; ok {
		projectId = fmt.Sprintf("%v", v)
	}
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	result, err := h.executionHandler.GetExecutionsHTTP(projectId, productId)
	if err != nil {
		h.sendErrorResponse(encoder, err.Error())
		return
	}
	h.sendResponse(encoder, map[string]interface{}{
		"status":  "ok",
		"message": "Executions retrieved successfully",
		"data":    result,
	})
}

func (h *MCPHandler) handleGetBugs(encoder *json.Encoder, params map[string]interface{}) {
	productId := ""
	status := ""
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	if v, ok := params["status"]; ok {
		status = fmt.Sprintf("%v", v)
	}
	result, err := h.bugHandler.GetBugsHTTP(productId, status)
	if err != nil {
		h.sendErrorResponse(encoder, err.Error())
		return
	}
	h.sendResponse(encoder, map[string]interface{}{
		"status":  "ok",
		"message": "Bugs retrieved successfully",
		"data":    result,
	})
}

func (h *MCPHandler) handleGetStories(encoder *json.Encoder, params map[string]interface{}) {
	productId := ""
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	result, err := h.storyHandler.GetStoriesHTTP(productId)
	if err != nil {
		h.sendErrorResponse(encoder, err.Error())
		return
	}
	h.sendResponse(encoder, map[string]interface{}{
		"status":  "ok",
		"message": "Stories retrieved successfully",
		"data":    result,
	})
}

func (h *MCPHandler) handleGetTasks(encoder *json.Encoder, params map[string]interface{}) {
	productId := ""
	executionId := ""
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	if v, ok := params["executionId"]; ok {
		executionId = fmt.Sprintf("%v", v)
	}
	result, err := h.taskHandler.GetTasksHTTP(productId, executionId)
	if err != nil {
		h.sendErrorResponse(encoder, err.Error())
		return
	}
	h.sendResponse(encoder, map[string]interface{}{
		"status":  "ok",
		"message": "Tasks retrieved successfully",
		"data":    result,
	})
}

func (h *MCPHandler) handleGetUsers(encoder *json.Encoder, params map[string]interface{}) {
	result, err := h.userHandler.GetUsersHTTP()
	if err != nil {
		h.sendErrorResponse(encoder, err.Error())
		return
	}
	h.sendResponse(encoder, map[string]interface{}{
		"status":  "ok",
		"message": "Users retrieved successfully",
		"data":    result,
	})
}

func (h *MCPHandler) handleGetTimelog(encoder *json.Encoder, params map[string]interface{}) {
	productId := ""
	dateFrom := ""
	dateTo := ""
	if v, ok := params["productId"]; ok {
		productId = fmt.Sprintf("%v", v)
	}
	if v, ok := params["dateFrom"]; ok {
		dateFrom = fmt.Sprintf("%v", v)
	}
	if v, ok := params["dateTo"]; ok {
		dateTo = fmt.Sprintf("%v", v)
	}
	result, err := h.timelogHandler.GetTimelogHTTP(productId, dateFrom, dateTo)
	if err != nil {
		h.sendErrorResponse(encoder, err.Error())
		return
	}
	h.sendResponse(encoder, map[string]interface{}{
		"status":  "ok",
		"message": "Timelog retrieved successfully",
		"data":    result,
	})
}

func (h *MCPHandler) handlePing(encoder *json.Encoder) {
	h.sendResponse(encoder, map[string]interface{}{
		"status":  "ok",
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
		"status":  "error",
		"message": errorMsg,
	})
}
