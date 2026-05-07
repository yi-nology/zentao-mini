package mcp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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

// HTTPTransport HTTP 传输层
type HTTPTransport struct {
	server *MCPServer
}

// NewHTTPTransport 创建 HTTP 传输层
func NewHTTPTransport(server *MCPServer) *HTTPTransport {
	return &HTTPTransport{server: server}
}

// respond 辅助：统一返回格式
func respond(c *gin.Context, result interface{}, err error) {
	if err != nil {
		c.JSON(http.StatusBadRequest, MCPResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	if m, ok := result.(map[string]interface{}); ok {
		c.JSON(http.StatusOK, m)
		return
	}
	c.JSON(http.StatusOK, MCPResponse{
		Status: "ok",
		Data:   result,
	})
}

// collectQueryParams 从 URL query 收集已知参数
func collectQueryParams(c *gin.Context) map[string]interface{} {
	params := make(map[string]interface{})
	for _, key := range []string{"productId", "projectId", "executionId", "status", "assignedTo", "dateFrom", "dateTo", "page", "pageSize"} {
		if val := c.Query(key); val != "" {
			params[key] = val
		}
	}
	return params
}

// HandleAction POST /mcp — 统一 JSON 入口
func (t *HTTPTransport) HandleAction(c *gin.Context) {
	var req MCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MCPResponse{
			Status:  "error",
			Message: "Invalid request: expecting JSON {\"action\":\"...\",\"params\":{...}}",
		})
		return
	}
	result, err := t.server.HandleAction(req.Action, req.Params)
	respond(c, result, err)
}

// HandleActionGet GET /mcp?action=xxx — 查询参数方式
func (t *HTTPTransport) HandleActionGet(c *gin.Context) {
	action := c.Query("action")
	if action == "" {
		c.JSON(http.StatusBadRequest, MCPResponse{
			Status:  "error",
			Message: "Usage: GET /mcp?action=<action_name>&param=value  OR  GET /mcp/<action>?param=value",
		})
		return
	}
	result, err := t.server.HandleAction(action, collectQueryParams(c))
	respond(c, result, err)
}

// HandleListTools GET /mcp/tools — 列出所有可用 tools
func (t *HTTPTransport) HandleListTools(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"count":  len(Tools),
		"tools":  Tools,
	})
}

// HandleGetTool GET /mcp/tools/:name — 获取单个 tool 详情
func (t *HTTPTransport) HandleGetTool(c *gin.Context) {
	name := c.Param("name")
	tool := GetToolByName(name)
	if tool == nil {
		c.JSON(http.StatusNotFound, MCPResponse{
			Status:  "error",
			Message: "Tool not found: " + name,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"tool":   tool,
	})
}

// RegisterRoutes 注册所有 MCP HTTP 路由到 gin.Engine
// 注意：必须先注册具体路径，再注册通配路径
func (t *HTTPTransport) RegisterRoutes(r *gin.Engine) {
	// ===== Tools 发现端点（最具体，优先匹配） =====
	r.GET("/mcp/tools", t.HandleListTools)
	r.GET("/mcp/tools/:name", t.HandleGetTool)

	// ===== 便捷端点：GET /mcp/<action> =====
	// 这些路径比 /mcp 更具体，Gin 的 radix tree 会优先匹配
	actions := []string{"ping", "products", "projects", "executions", "bugs", "stories", "tasks", "users", "timelog"}
	for _, a := range actions {
		action := a // capture
		// GET /mcp/<action>?param=value
		r.GET("/mcp/"+action, func(c *gin.Context) {
			result, err := t.server.HandleAction(action, collectQueryParams(c))
			respond(c, result, err)
		})
		// POST /mcp/<action> — 也支持 JSON body
		r.POST("/mcp/"+action, func(c *gin.Context) {
			var req MCPRequest
			if err := c.ShouldBindJSON(&req); err == nil && req.Action != "" {
				// 有 JSON body，用 body 里的 action
				result, err := t.server.HandleAction(req.Action, req.Params)
				respond(c, result, err)
			} else {
				// 没有 JSON body，用 URL 路径推断 action
				result, err := t.server.HandleAction(action, collectQueryParams(c))
				respond(c, result, err)
			}
		})
	}

	// ===== 统一入口（最后注册，优先级最低） =====
	r.POST("/mcp", t.HandleAction)
	r.GET("/mcp", t.HandleActionGet)
}
