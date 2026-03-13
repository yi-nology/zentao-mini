package handlers

import (
	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/dto"
	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/service"
)

// BugHandler Bug处理器
// 只负责HTTP请求/响应处理，业务逻辑由Service层处理
type BugHandler struct {
	bugService *service.BugService
}

// NewBugHandler 创建Bug处理器
func NewBugHandler(bugService *service.BugService) *BugHandler {
	return &BugHandler{bugService: bugService}
}

// GetBugs 获取Bug列表
// @Summary 获取Bug列表
// @Description 获取Bug列表，支持按产品、项目、状态、指派人、时间范围筛选
// @Tags Bug
// @Accept json
// @Produce json
// @Param productId query int false "产品ID"
// @Param projectId query int false "项目ID"
// @Param status query string false "状态(active, resolved, closed等)"
// @Param assignedTo query string false "指派人账号"
// @Param startDate query string false "开始日期(YYYY-MM-DD)"
// @Param endDate query string false "结束日期(YYYY-MM-DD)"
// @Param specificDate query string false "具体日期(YYYY-MM-DD)"
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认20"
// @Success 200 {object} errors.Response{data=vo.PaginatedVO{list=[]vo.BugVO}}
// @Failure 400 {object} errors.Response
// @Failure 500 {object} errors.Response
// @Router /api/v1/bugs [get]
func (h *BugHandler) GetBugs(c *gin.Context) {
	// 绑定请求参数到DTO
	var query dto.BugQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.BadRequest(c, "参数格式错误")
		return
	}

	// 验证参数
	if err := query.Validate(); err != nil {
		errors.Error(c, err)
		return
	}

	// 调用Service层处理业务逻辑
	result, err := h.bugService.GetBugs(&query)
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, result)
}
