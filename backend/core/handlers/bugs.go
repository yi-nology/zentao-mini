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

// BugHandler Bug处理器
type BugHandler struct {
	client *myzentao.Client
}

// NewBugHandler 创建Bug处理器
func NewBugHandler(client *myzentao.Client) *BugHandler {
	return &BugHandler{client: client}
}

// GetBugs 获取Bug列表
// @Summary 获取Bug列表
// @Description 获取Bug列表，支持按产品、项目、状态、指派人、时间范围筛选
// @Tags Bug
// @Accept json
// @Produce json
// @Param productID query int false "产品ID"
// @Param projectID query int false "项目ID"
// @Param status query string false "状态(active, resolved, closed等)"
// @Param assignedTo query string false "指派人账号"
// @Param startDate query string false "开始日期(YYYY-MM-DD)"
// @Param endDate query string false "结束日期(YYYY-MM-DD)"
// @Param specificDate query string false "具体日期(YYYY-MM-DD)"
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认20"
// @Success 200 {object} models.Response{data=models.PaginatedResult{list=[]models.Bug}}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/bugs [get]
func (h *BugHandler) GetBugs(c *gin.Context) {
	productIDStr := c.Query("productID")
	projectIDStr := c.Query("projectID")
	status := c.Query("status")
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

	var bugs []zentao.Bug

	// 如果有产品ID，按产品查询
	if productIDStr != "" {
		productID, err := strconv.Atoi(productIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "无效的产品ID",
			})
			return
		}

		// 优先使用SearchBugs进行多条件筛选，减少内存消耗
		if assignedTo != "" || status != "" {
			params := zentao.BugSearchParams{
				ProductID:  productID,
				Status:     status,
				AssignedTo: assignedTo,
				Limit:      1000, // 一次获取足够多的数据用于筛选
				Page:       1,
			}
			bugs, err = h.client.SearchBugs(params)
		} else if projectIDStr != "" {
			// 如果只有项目ID，使用GetBugsByProject
			projectID, err := strconv.Atoi(projectIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.Response{
					Code:    400,
					Message: "无效的项目ID",
				})
				return
			}
			bugs, err = h.client.GetBugsByProject(productID, projectID, 1000)
		} else {
			// 获取产品的所有Bug
			bugs, err = h.client.GetBugs(productID, 1000)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "获取Bug列表失败: " + err.Error(),
			})
			return
		}
	} else {
		// 如果没有产品ID，返回空列表
		bugs = []zentao.Bug{}
	}

	// 按时间范围筛选（时间范围筛选需要在客户端进行，因为API不支持）
	if startDate != "" || endDate != "" || specificDate != "" {
		var filteredBugs []zentao.Bug
		for _, b := range bugs {
			bugDate := b.OpenedDate
			// 如果有具体日期，优先按具体日期筛选
			if specificDate != "" {
				if strings.HasPrefix(bugDate, specificDate) {
					filteredBugs = append(filteredBugs, b)
				}
			} else if startDate != "" || endDate != "" {
				// 按时间范围筛选
				if (startDate == "" || bugDate >= startDate) && (endDate == "" || bugDate <= endDate) {
					filteredBugs = append(filteredBugs, b)
				}
			}
		}
		bugs = filteredBugs
	}

	// 计算总记录数
	total := len(bugs)

	// 计算分页
	start := (page - 1) * limit
	end := start + limit
	if start >= total {
		// 超出范围，返回空列表
		c.JSON(http.StatusOK, models.Response{
			Code:    200,
			Message: "success",
			Data: models.PaginatedResult{
				List:  []models.Bug{},
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
	pagedBugs := bugs[start:end]

	// 转换为模型
	result := make([]models.Bug, 0)
	for _, b := range pagedBugs {
		result = append(result, convertBug(b))
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

// convertBug 转换Bug模型
func convertBug(b zentao.Bug) models.Bug {
	return models.Bug{
		ID:            b.ID,
		Project:       b.Project,
		Product:       b.Product,
		Title:         b.Title,
		Keywords:      b.Keywords,
		Severity:      b.Severity,
		Pri:           b.Pri,
		Type:          b.Type,
		OS:            b.OS,
		Browser:       b.Browser,
		Hardware:      b.Hardware,
		Steps:         b.Steps,
		Status:        b.Status,
		SubStatus:     b.SubStatus,
		Color:         b.Color,
		Confirmed:     b.Confirmed,
		PlanTime:      b.PlanTime,
		OpenedBy:      models.UserRef(b.OpenedBy),
		OpenedDate:    b.OpenedDate,
		OpenedBuild:   b.OpenedBuild,
		AssignedTo:    models.UserRef(b.AssignedTo),
		AssignedDate:  b.AssignedDate,
		Deadline:      b.Deadline,
		ResolvedBy:    b.ResolvedBy,
		Resolution:    b.Resolution,
		ResolvedBuild: b.ResolvedBuild,
		ResolvedDate:  b.ResolvedDate,
		ClosedBy:      b.ClosedBy,
		ClosedDate:    b.ClosedDate,
		StatusName:    b.StatusName,
		LifeCycle:     b.LifeCycle,
	}
}
