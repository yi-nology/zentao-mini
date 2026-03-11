package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yi-nology/common/biz/zentao"

	"chandao-mini/backend/core/models"
	myzentao "chandao-mini/backend/core/zentao"
)

// StoryHandler 需求处理器
type StoryHandler struct {
	client *myzentao.Client
}

// NewStoryHandler 创建需求处理器
func NewStoryHandler(client *myzentao.Client) *StoryHandler {
	return &StoryHandler{client: client}
}

// GetStories 获取需求列表
// @Summary 获取需求列表
// @Description 获取需求列表，支持按产品、项目、执行、指派人、时间范围筛选
// @Tags 需求
// @Accept json
// @Produce json
// @Param productID query int false "产品ID"
// @Param projectID query int false "项目ID"
// @Param executionID query int false "执行ID(迭代ID)"
// @Param assignedTo query string false "指派人账号或姓名"
// @Param startDate query string false "开始日期(YYYY-MM-DD)"
// @Param endDate query string false "结束日期(YYYY-MM-DD)"
// @Param specificDate query string false "具体日期(YYYY-MM-DD)"
// @Param page query int false "页码，默认1"
// @Param pageSize query int false "每页数量，默认20"
// @Param limit query int false "每页数量，默认20"
// @Success 200 {object} models.Response{data=models.PaginatedResult{list=[]models.Story}}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/stories [get]
func (h *StoryHandler) GetStories(c *gin.Context) {
	productIDStr := c.Query("productID")
	projectIDStr := c.Query("projectID")
	executionIDStr := c.Query("executionID")
	assignedTo := c.Query("assignedTo")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	specificDate := c.Query("specificDate")

	// 分页参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("pageSize", c.DefaultQuery("limit", "20"))
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	var stories []zentao.Story

	// 优先级: executionID > projectID > productID
	if executionIDStr != "" {
		executionID, err := strconv.Atoi(executionIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "无效的执行ID",
			})
			return
		}
		stories, err = h.client.GetStoriesByExecution(executionID, 500)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "获取需求列表失败: " + err.Error(),
			})
			return
		}
	} else if projectIDStr != "" {
		projectID, err := strconv.Atoi(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "无效的项目ID",
			})
			return
		}
		stories, err = h.client.GetStoriesByProject(projectID, 500)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "获取需求列表失败: " + err.Error(),
			})
			return
		}
	} else if productIDStr != "" {
		productID, err := strconv.Atoi(productIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "无效的产品ID",
			})
			return
		}
		stories, err = h.client.GetStoriesByProduct(productID, 500)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "获取需求列表失败: " + err.Error(),
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请提供产品ID、项目ID或执行ID",
		})
		return
	}

	// 按指派人筛选
	if assignedTo != "" {
		var filteredStories []zentao.Story
		for _, s := range stories {
			assignedStr := ""
			if s.AssignedTo != nil {
				switch v := s.AssignedTo.(type) {
				case string:
					assignedStr = v
				case map[string]interface{}:
					if account, ok := v["account"].(string); ok {
						assignedStr = account
					}
				}
			}
			if strings.EqualFold(assignedStr, assignedTo) {
				filteredStories = append(filteredStories, s)
			}
		}
		stories = filteredStories
	}

	// 按时间范围筛选
	if startDate != "" || endDate != "" || specificDate != "" {
		var filteredStories []zentao.Story
		for _, s := range stories {
			storyDate := s.OpenedDate
			// 如果有具体日期，优先按具体日期筛选
			if specificDate != "" {
				if strings.HasPrefix(storyDate, specificDate) {
					filteredStories = append(filteredStories, s)
				}
			} else if startDate != "" || endDate != "" {
				// 按时间范围筛选
				if (startDate == "" || storyDate >= startDate) && (endDate == "" || storyDate <= endDate) {
					filteredStories = append(filteredStories, s)
				}
			}
		}
		stories = filteredStories
	}

	// 计算总记录数
	total := len(stories)

	// 计算分页
	start := (page - 1) * limit
	end := start + limit
	if start >= total {
		// 超出范围，返回空列表
		c.JSON(http.StatusOK, models.Response{
			Code:    200,
			Message: "success",
			Data: models.PaginatedResult{
				List:  []models.Story{},
				Total: total,
				Page:  page,
				Limit: limit,
			},
		})
		return
	}
	if end > total {
		end = total
	}

	// 截取分页数据
	pagedStories := stories[start:end]

	// 转换为模型
	var result []models.Story
	for _, s := range pagedStories {
		result = append(result, models.Story{
			ID:           s.ID,
			Product:      s.Product,
			Module:       s.Module,
			Plan:         s.Plan,
			Source:       s.Source,
			Title:        s.Title,
			Spec:         s.Spec,
			Verify:       s.Verify,
			Type:         s.Type,
			Status:       s.Status,
			Stage:        s.Stage,
			Pri:          s.Pri,
			Estimate:     s.Estimate,
			Version:      s.Version,
			OpenedBy:     s.OpenedBy,
			OpenedDate:   s.OpenedDate,
			AssignedTo:   s.AssignedTo,
			AssignedDate: s.AssignedDate,
			ClosedBy:     s.ClosedBy,
			ClosedDate:   s.ClosedDate,
			ClosedReason: s.ClosedReason,
		})
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "success",
		Data: models.PaginatedResult{
			List:  result,
			Total: total,
			Page:  page,
			Limit: limit,
		},
	})
}
