package handlers

import (
	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/dto"
	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/service"
)

// TimelogHandler 工时统计处理器
// 只负责HTTP请求/响应处理，业务逻辑由Service层处理
type TimelogHandler struct {
	timelogService *service.TimelogService
}

// NewTimelogHandler 创建工时统计处理器
func NewTimelogHandler(timelogService *service.TimelogService) *TimelogHandler {
	return &TimelogHandler{timelogService: timelogService}
}

// GetTimelogAnalysis 获取工时统计分析
// @Summary 获取工时统计分析
// @Description 获取工时统计分析数据
// @Tags 工时统计
// @Accept json
// @Produce json
// @Param productId query string true "产品ID"
// @Param projectId query string false "项目ID"
// @Param executionId query string false "执行ID"
// @Param assignedTo query string false "指派人"
// @Param dateFrom query string true "开始日期(YYYY-MM-DD)"
// @Param dateTo query string true "结束日期(YYYY-MM-DD)"
// @Success 200 {object} errors.Response
// @Failure 400 {object} errors.Response
// @Failure 500 {object} errors.Response
// @Router /api/v1/timelog/analysis [get]
func (h *TimelogHandler) GetTimelogAnalysis(c *gin.Context) {
	// 绑定请求参数到DTO
	var query dto.TimelogQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	// 验证必要参数
	if query.ProductID == "" {
		errors.BadRequest(c, "productId is required")
		return
	}

	if query.DateFrom == "" || query.DateTo == "" {
		errors.BadRequest(c, "dateFrom and dateTo are required")
		return
	}

	// 调用Service层处理业务逻辑
	result, err := h.timelogService.GetTimelogAnalysis(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, result)
}

// GetTimelogDashboard 获取工时统计看板数据
// @Summary 获取工时统计看板数据
// @Description 获取工时统计看板数据
// @Tags 工时统计
// @Accept json
// @Produce json
// @Param productId query string true "产品ID"
// @Param projectId query string false "项目ID"
// @Param executionId query string false "执行ID"
// @Param assignedTo query string false "指派人"
// @Param dateFrom query string true "开始日期(YYYY-MM-DD)"
// @Param dateTo query string true "结束日期(YYYY-MM-DD)"
// @Success 200 {object} errors.Response
// @Failure 400 {object} errors.Response
// @Failure 500 {object} errors.Response
// @Router /api/v1/timelog/dashboard [get]
func (h *TimelogHandler) GetTimelogDashboard(c *gin.Context) {
	// 绑定请求参数到DTO
	var query dto.TimelogQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	// 验证必要参数
	if query.ProductID == "" {
		errors.BadRequest(c, "productId is required")
		return
	}

	if query.DateFrom == "" || query.DateTo == "" {
		errors.BadRequest(c, "dateFrom and dateTo are required")
		return
	}

	// 调用Service层处理业务逻辑
	result, err := h.timelogService.GetTimelogDashboard(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, result)
}

// GetTimelogEfforts 获取工时流水明细
// @Summary 获取工时流水明细
// @Description 获取工时流水明细数据
// @Tags 工时统计
// @Accept json
// @Produce json
// @Param productId query string true "产品ID"
// @Param projectId query string false "项目ID"
// @Param executionId query string false "执行ID"
// @Param assignedTo query string false "指派人"
// @Param dateFrom query string true "开始日期(YYYY-MM-DD)"
// @Param dateTo query string true "结束日期(YYYY-MM-DD)"
// @Success 200 {object} errors.Response
// @Failure 400 {object} errors.Response
// @Failure 500 {object} errors.Response
// @Router /api/v1/timelog/efforts [get]
func (h *TimelogHandler) GetTimelogEfforts(c *gin.Context) {
	// 绑定请求参数到DTO
	var query dto.TimelogQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	// 验证必要参数
	if query.ProductID == "" {
		errors.BadRequest(c, "productId is required")
		return
	}

	if query.DateFrom == "" || query.DateTo == "" {
		errors.BadRequest(c, "dateFrom and dateTo are required")
		return
	}

	// 调用Service层处理业务逻辑
	result, err := h.timelogService.GetTimelogEfforts(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, result)
}
