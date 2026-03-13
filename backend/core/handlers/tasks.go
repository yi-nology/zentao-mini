package handlers

import (
	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/dto"
	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/service"
)

// TaskHandler 任务处理器
// 只负责HTTP请求/响应处理，业务逻辑由Service层处理
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// GetTasks 获取任务列表
// @Summary 获取任务列表
// @Description 获取执行(迭代)下的任务列表，支持按人、状态、时间筛选
// @Tags 任务
// @Accept json
// @Produce json
// @Param executionId query int true "执行ID(迭代ID)"
// @Param assignedTo query string false "指派人"
// @Param status query string false "任务状态"
// @Param startDate query string false "开始日期 (YYYY-MM-DD)"
// @Param endDate query string false "结束日期 (YYYY-MM-DD)"
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认20"
// @Success 200 {object} errors.Response{data=vo.PaginatedVO{list=[]vo.TaskVO}}
// @Failure 400 {object} errors.Response
// @Failure 500 {object} errors.Response
// @Router /api/v1/tasks [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
	// 绑定请求参数到DTO
	var query dto.TaskQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	// 验证必需参数
	if query.ExecutionID == 0 {
		errors.MissingParam(c, "executionId")
		return
	}

	// 验证参数
	if err := query.Validate(); err != nil {
		errors.Error(c, err)
		return
	}

	// 调用Service层处理业务逻辑
	result, err := h.taskService.GetTasks(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, result)
}
