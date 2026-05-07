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

// HandleAction POST /mcp — 统一 JSON 入口
func (t *HTTPTransport) HandleAction(c *gin.Context) {
	var req MCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MCPResponse{
			Status:  "error",
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := t.server.HandleAction(req.Action, req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, MCPResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// result 可能已经是包含 status 的 map，直接返回
	if m, ok := result.(map[string]interface{}); ok {
		c.JSON(http.StatusOK, m)
		return
	}

	c.JSON(http.StatusOK, MCPResponse{
		Status: "ok",
		Data:   result,
	})
}

// HandleActionGet GET /mcp?action=xxx — 查询参数方式
func (t *HTTPTransport) HandleActionGet(c *gin.Context) {
	action := c.Query("action")
	if action == "" {
		c.JSON(http.StatusBadRequest, MCPResponse{
			Status:  "error",
			Message: "Missing 'action' query parameter",
		})
		return
	}

	params := make(map[string]interface{})
	for _, key := range []string{"productId", "projectId", "executionId", "status", "assignedTo", "dateFrom", "dateTo", "page", "pageSize"} {
		if val := c.Query(key); val != "" {
			params[key] = val
		}
	}

	result, err := t.server.HandleAction(action, params)
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
func (t *HTTPTransport) RegisterRoutes(r *gin.Engine) {
	// 统一入口
	r.POST("/mcp", t.HandleAction)
	r.GET("/mcp", t.HandleActionGet)

	// Tools 发现端点
	r.GET("/mcp/tools", t.HandleListTools)
	r.GET("/mcp/tools/:name", t.HandleGetTool)

	// 便捷端点 - POST
	r.POST("/mcp/ping", func(c *gin.Context) { t.HandleAction(c) })
	r.POST("/mcp/products", func(c *gin.Context) { t.HandleAction(c) })
	r.POST("/mcp/projects", func(c *gin.Context) { t.HandleAction(c) })
	r.POST("/mcp/executions", func(c *gin.Context) { t.HandleAction(c) })
	r.POST("/mcp/bugs", func(c *gin.Context) { t.HandleAction(c) })
	r.POST("/mcp/stories", func(c *gin.Context) { t.HandleAction(c) })
	r.POST("/mcp/tasks", func(c *gin.Context) { t.HandleAction(c) })
	r.POST("/mcp/users", func(c *gin.Context) { t.HandleAction(c) })
	r.POST("/mcp/timelog", func(c *gin.Context) { t.HandleAction(c) })

	// 便捷端点 - GET
	r.GET("/mcp/ping", func(c *gin.Context) {
		c.Request.URL.RawQuery = "action=ping"
		t.HandleActionGet(c)
	})
	r.GET("/mcp/products", func(c *gin.Context) {
		c.Request.URL.RawQuery = "action=get_products"
		t.HandleActionGet(c)
	})
	r.GET("/mcp/bugs", func(c *gin.Context) {
		c.Request.URL.RawQuery = "action=get_bugs&" + c.Request.URL.RawQuery
		t.HandleActionGet(c)
	})
	r.GET("/mcp/stories", func(c *gin.Context) {
		c.Request.URL.RawQuery = "action=get_stories&" + c.Request.URL.RawQuery
		t.HandleActionGet(c)
	})
	r.GET("/mcp/tasks", func(c *gin.Context) {
		c.Request.URL.RawQuery = "action=get_tasks&" + c.Request.URL.RawQuery
		t.HandleActionGet(c)
	})
	r.GET("/mcp/users", func(c *gin.Context) {
		c.Request.URL.RawQuery = "action=get_users"
		t.HandleActionGet(c)
	})
}
