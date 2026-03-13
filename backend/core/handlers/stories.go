package handlers

import (
	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/dto"
	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/service"
)

// StoryHandler 需求处理器
// 只负责HTTP请求/响应处理，业务逻辑由Service层处理
type StoryHandler struct {
	storyService *service.StoryService
}

// NewStoryHandler 创建需求处理器
func NewStoryHandler(storyService *service.StoryService) *StoryHandler {
	return &StoryHandler{storyService: storyService}
}

// GetStories 获取需求列表
// @Summary 获取需求列表
// @Description 获取需求列表，支持按产品、项目、执行、指派人、时间范围筛选
// @Tags 需求
// @Accept json
// @Produce json
// @Param productId query int false "产品ID"
// @Param projectId query int false "项目ID"
// @Param executionId query int false "执行ID(迭代ID)"
// @Param assignedTo query string false "指派人账号或姓名"
// @Param startDate query string false "开始日期(YYYY-MM-DD)"
// @Param endDate query string false "结束日期(YYYY-MM-DD)"
// @Param specificDate query string false "具体日期(YYYY-MM-DD)"
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认20"
// @Success 200 {object} errors.Response{data=vo.PaginatedVO{list=[]vo.StoryVO}}
// @Failure 400 {object} errors.Response
// @Failure 500 {object} errors.Response
// @Router /api/v1/stories [get]
func (h *StoryHandler) GetStories(c *gin.Context) {
	// 绑定请求参数到DTO
	var query dto.StoryQueryDTO
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
	result, err := h.storyService.GetStories(&query)
	if err != nil {
		// 处理验证错误
		if _, ok := err.(*service.ValidationError); ok {
			errors.BadRequest(c, err.Error())
			return
		}
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, result)
}
