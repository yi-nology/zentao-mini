package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/models"
	myzentao "chandao-mini/backend/core/zentao"

	"github.com/yi-nology/common/biz/zentao"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	client *myzentao.Client
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(client *myzentao.Client) *TaskHandler {
	return &TaskHandler{client: client}
}

// GetTasks 获取任务列表
// @Summary 获取任务列表
// @Description 获取执行(迭代)下的任务列表，支持按人、状态、时间筛选
// @Tags 任务
// @Accept json
// @Produce json
// @Param executionID query int true "执行ID(迭代ID)"
// @Param assignedTo query string false "指派人"
// @Param status query string false "任务状态"
// @Param startDate query string false "开始日期 (YYYY-MM-DD)"
// @Param endDate query string false "结束日期 (YYYY-MM-DD)"
// @Param page query int false "页码，默认1"
// @Param pageSize query int false "每页数量，默认20"
// @Param limit query int false "每页数量，默认20"
// @Success 200 {object} models.Response{data=models.PaginatedResult{list=[]models.Task}}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/tasks [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
	executionIDStr := c.Query("executionID")

	if executionIDStr == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请提供执行ID",
		})
		return
	}

	executionID, err := strconv.Atoi(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "无效的执行ID",
		})
		return
	}

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

	// 获取筛选参数
	assignedToStr := c.Query("assignedTo")
	status := c.Query("status")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// 转换指派人ID为整数
	var assignedToID int
	if assignedToStr != "" {
		var err error
		assignedToID, err = strconv.Atoi(assignedToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "无效的指派人ID",
			})
			return
		}
	}

	tasks, err := h.client.GetTasks(executionID, 500)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "获取任务列表失败: " + err.Error(),
		})
		return
	}

	// 过滤任务
	var filteredTasks []zentao.Task
	for _, t := range tasks {
		// 按指派人筛选
		if assignedToStr != "" {
			if t.AssignedTo.ID == 0 || t.AssignedTo.ID != assignedToID {
				continue
			}
		}

		// 按状态筛选
		if status != "" {
			if t.Status != status {
				continue
			}
		}

		// 按日期范围筛选（使用创建日期）
		if startDate != "" || endDate != "" {
			taskDate := t.OpenedDate
			if taskDate == "" {
				continue
			}

			// 提取日期部分（去除时间）
			if len(taskDate) > 10 {
				taskDate = taskDate[:10]
			}

			if startDate != "" && taskDate < startDate {
				continue
			}
			if endDate != "" && taskDate > endDate {
				continue
			}
		}

		filteredTasks = append(filteredTasks, t)
	}

	// 计算总记录数
	total := len(filteredTasks)

	// 计算分页
	start := (page - 1) * limit
	end := start + limit
	if start >= total {
		// 超出范围，返回空列表
		c.JSON(http.StatusOK, models.Response{
			Code:    200,
			Message: "success",
			Data: models.PaginatedResult{
				List:  []models.Task{},
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
	pagedTasks := filteredTasks[start:end]

	// 转换为模型
	var result []models.Task
	for _, t := range pagedTasks {
		result = append(result, models.Task{
			ID:           t.ID,
			Project:      t.Project,
			Execution:    t.Execution,
			Name:         t.Name,
			Type:         t.Type,
			Pri:          t.Pri,
			Status:       t.Status,
			AssignedTo:   models.UserRef(t.AssignedTo),
			EstStarted:   t.EstStarted,
			Deadline:     t.Deadline,
			Estimate:     t.Estimate,
			Consumed:     t.Consumed,
			Left:         t.Left,
			Desc:         t.Desc,
			OpenedBy:     t.OpenedBy,
			OpenedDate:   t.OpenedDate,
			FinishedBy:   t.FinishedBy,
			FinishedDate: t.FinishedDate,
			ClosedBy:     t.ClosedBy,
			ClosedDate:   t.ClosedDate,
			StatusName:   t.StatusName,
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
