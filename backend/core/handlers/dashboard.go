package handlers

import (
	"chandao-mini/backend/core/dto"
	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/service"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(ds *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: ds}
}

// GetDashboard 获取仪表盘数据
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	var query dto.DashboardQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}
	result, err := h.dashboardService.GetDashboardContext(c.Request.Context(), query.ProductID)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}
	errors.Success(c, result)
}

// GetProjectOverview 获取项目概览
func (h *DashboardHandler) GetProjectOverview(c *gin.Context) {
	var query dto.ProjectOverviewQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}
	result, err := h.dashboardService.GetProjectOverviewContext(c.Request.Context(), query.ProjectID)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}
	errors.Success(c, result)
}

// GetPersonalTimelog 获取个人工时报表
func (h *DashboardHandler) GetPersonalTimelog(c *gin.Context) {
	var query dto.PersonalTimelogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}
	result, err := h.dashboardService.GetPersonalTimelogContext(c.Request.Context(), query.Account, query.ProductID, query.DateFrom, query.DateTo, query.GroupBy)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}
	errors.Success(c, result)
}

// Search 全局搜索
func (h *DashboardHandler) Search(c *gin.Context) {
	var query dto.SearchQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}
	if err := query.Validate(); err != nil {
		errors.Error(c, err)
		return
	}
	result, err := h.dashboardService.SearchContext(c.Request.Context(), query.Keyword, query.ProductID, query.Page, query.PageSize)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}
	errors.Success(c, result)
}
