package handlers

// Deprecated: 此文件中的 HTTP 辅助方法已被 backend/core/mcp/handlers.go 替代
// 新代码请使用 mcp.MCPServer 直接调用 service 层

import (
	"chandao-mini/backend/core/dto"
	"strconv"
)

// ============================================
// ProductHandler HTTP 方法
// ============================================

// GetProductsHTTP 获取产品列表（HTTP模式）
func (h *ProductHandler) GetProductsHTTP() (interface{}, error) {
	return h.productService.GetProducts()
}

// ============================================
// ProjectHandler HTTP 方法
// ============================================

// GetProjectsHTTP 获取项目列表（HTTP模式）
func (h *ProjectHandler) GetProjectsHTTP(productId string) (interface{}, error) {
	query := &dto.ProjectQueryDTO{}
	if productId != "" {
		if v, err := strconv.Atoi(productId); err == nil {
			query.ProductID = v
		}
	}
	return h.projectService.GetProjects(query)
}

// ============================================
// ExecutionHandler HTTP 方法
// ============================================

// GetExecutionsHTTP 获取执行列表（HTTP模式）
func (h *ExecutionHandler) GetExecutionsHTTP(projectId, productId string) (interface{}, error) {
	query := &dto.ExecutionQueryDTO{}
	if projectId != "" {
		if v, err := strconv.Atoi(projectId); err == nil {
			query.ProjectID = v
		}
	}
	if productId != "" {
		if v, err := strconv.Atoi(productId); err == nil {
			query.ProductID = v
		}
	}
	return h.executionService.GetExecutions(query)
}

// ============================================
// BugHandler HTTP 方法
// ============================================

// GetBugsHTTP 获取Bug列表（HTTP模式）
func (h *BugHandler) GetBugsHTTP(productId, status string) (interface{}, error) {
	query := &dto.BugQueryDTO{}
	if productId != "" {
		if v, err := strconv.Atoi(productId); err == nil {
			query.ProductID = v
		}
	}
	query.Status = status
	query.Page = 1
	query.PageSize = 100
	return h.bugService.GetBugs(query)
}

// ============================================
// StoryHandler HTTP 方法
// ============================================

// GetStoriesHTTP 获取需求列表（HTTP模式）
func (h *StoryHandler) GetStoriesHTTP(productId string) (interface{}, error) {
	query := &dto.StoryQueryDTO{}
	if productId != "" {
		if v, err := strconv.Atoi(productId); err == nil {
			query.ProductID = v
		}
	}
	query.Page = 1
	query.PageSize = 100
	return h.storyService.GetStories(query)
}

// ============================================
// TaskHandler HTTP 方法
// ============================================

// GetTasksHTTP 获取任务列表（HTTP模式）
func (h *TaskHandler) GetTasksHTTP(productId, executionId string) (interface{}, error) {
	query := &dto.TaskQueryDTO{}
	if productId != "" {
		if v, err := strconv.Atoi(productId); err == nil {
			query.ProductID = v
		}
	}
	if executionId != "" {
		if v, err := strconv.Atoi(executionId); err == nil {
			query.ExecutionID = v
		}
	}
	query.Page = 1
	query.PageSize = 100
	return h.taskService.GetTasks(query)
}

// ============================================
// UserHandler HTTP 方法
// ============================================

// GetUsersHTTP 获取用户列表（HTTP模式）
func (h *UserHandler) GetUsersHTTP() (interface{}, error) {
	return h.userService.GetUsers(1, 100)
}

// ============================================
// TimelogHandler HTTP 方法
// ============================================

// GetTimelogHTTP 获取工时数据（HTTP模式）
func (h *TimelogHandler) GetTimelogHTTP(productId, dateFrom, dateTo string) (interface{}, error) {
	query := &dto.TimelogQueryDTO{
		ProductID: productId,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
	}
	return h.timelogService.GetTimelogDashboard(query)
}
