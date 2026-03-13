package handlers

import (
	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/dto"
	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/service"
)

// ProjectHandler 项目处理器
// 只负责HTTP请求/响应处理，业务逻辑由Service层处理
type ProjectHandler struct {
	projectService *service.ProjectService
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

// GetProjects 获取项目列表
// @Summary 获取项目列表
// @Description 获取禅道系统中的所有项目列表，支持按产品筛选
// @Tags 项目
// @Accept json
// @Produce json
// @Param productId query int false "产品ID"
// @Success 200 {object} errors.Response{data=[]vo.ProjectVO}
// @Failure 500 {object} errors.Response
// @Router /api/v1/projects [get]
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	// 绑定请求参数到DTO
	var query dto.ProjectQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	// 调用Service层处理业务逻辑
	result, err := h.projectService.GetProjects(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, result)
}

// ExecutionHandler 执行/迭代处理器
// 只负责HTTP请求/响应处理，业务逻辑由Service层处理
type ExecutionHandler struct {
	executionService *service.ExecutionService
}

// NewExecutionHandler 创建执行/迭代处理器
func NewExecutionHandler(executionService *service.ExecutionService) *ExecutionHandler {
	return &ExecutionHandler{executionService: executionService}
}

// GetExecutions 获取执行列表
// @Summary 获取执行列表
// @Description 获取项目下的执行/迭代列表，支持按产品ID筛选
// @Tags 执行
// @Accept json
// @Produce json
// @Param projectId query int false "项目ID"
// @Param productId query int false "产品ID"
// @Success 200 {object} errors.Response{data=[]vo.ExecutionVO}
// @Failure 500 {object} errors.Response
// @Router /api/v1/executions [get]
func (h *ExecutionHandler) GetExecutions(c *gin.Context) {
	// 绑定请求参数到DTO
	var query dto.ExecutionQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	// 调用Service层处理业务逻辑
	result, err := h.executionService.GetExecutions(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, result)
}
