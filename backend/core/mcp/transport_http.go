package mcp

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

type MCPRequest struct {
	Action string                 `json:"action" binding:"required"`
	Params map[string]interface{} `json:"params"`
}

type MCPResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Version string      `json:"version,omitempty"`
}

type HTTPTransport struct {
	server *MCPServer
}

func NewHTTPTransport(server *MCPServer) *HTTPTransport {
	return &HTTPTransport{server: server}
}

func (t *HTTPTransport) Server() *MCPServer {
	return t.server
}

func (t *HTTPTransport) HandleActionByName(ctx context.Context, action string, c *app.RequestContext) {
	result, err := t.server.HandleAction(action, collectQueryParams(c))
	respond(c, result, err)
}

func (t *HTTPTransport) HandleActionByNameWithParams(ctx context.Context, action string, params map[string]interface{}, c *app.RequestContext) {
	result, err := t.server.HandleAction(action, params)
	respond(c, result, err)
}

func Respond(c *app.RequestContext, result interface{}, err error) {
	respond(c, result, err)
}

func respond(c *app.RequestContext, result interface{}, err error) {
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

func CollectQueryParams(c *app.RequestContext) map[string]interface{} {
	return collectQueryParams(c)
}

func collectQueryParams(c *app.RequestContext) map[string]interface{} {
	params := make(map[string]interface{})
	for _, key := range []string{"productId", "projectId", "executionId", "status", "assignedTo", "dateFrom", "dateTo", "page", "pageSize"} {
		if val := c.Query(key); val != "" {
			params[key] = val
		}
	}
	return params
}

func (t *HTTPTransport) HandleAction(ctx context.Context, c *app.RequestContext) {
	var req MCPRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, MCPResponse{
			Status:  "error",
			Message: "Invalid request: expecting JSON {\"action\":\"...\",\"params\":{...}}",
		})
		return
	}
	result, err := t.server.HandleAction(req.Action, req.Params)
	respond(c, result, err)
}

func (t *HTTPTransport) HandleActionGet(ctx context.Context, c *app.RequestContext) {
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

func (t *HTTPTransport) HandleListTools(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
		"count":  len(Tools),
		"tools":  Tools,
	})
}

func (t *HTTPTransport) HandleGetTool(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	tool := GetToolByName(name)
	if tool == nil {
		c.JSON(http.StatusNotFound, MCPResponse{
			Status:  "error",
			Message: "Tool not found: " + name,
		})
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
		"tool":   tool,
	})
}

func (t *HTTPTransport) RegisterRoutes(r *server.Hertz) {
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

	r.GET("/mcp/tools", t.HandleListTools)
	r.GET("/mcp/tools/:name", t.HandleGetTool)

	for path, action := range actionMap {
		act := action
		r.GET("/mcp/"+path, func(ctx context.Context, c *app.RequestContext) {
			result, err := t.server.HandleAction(act, collectQueryParams(c))
			respond(c, result, err)
		})
		r.POST("/mcp/"+path, func(ctx context.Context, c *app.RequestContext) {
			var req MCPRequest
			if err := c.BindAndValidate(&req); err == nil && req.Action != "" {
				result, err := t.server.HandleAction(req.Action, req.Params)
				respond(c, result, err)
			} else {
				result, err := t.server.HandleAction(act, collectQueryParams(c))
				respond(c, result, err)
			}
		})
	}

	r.POST("/mcp", t.HandleAction)
	r.GET("/mcp", t.HandleActionGet)
}
