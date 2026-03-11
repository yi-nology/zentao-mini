package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/models"
	"chandao-mini/backend/core/zentao"
)

// ExecutionHandler 执行/迭代处理器
type ExecutionHandler struct {
	client *zentao.Client
}

// NewExecutionHandler 创建执行/迭代处理器
func NewExecutionHandler(client *zentao.Client) *ExecutionHandler {
	return &ExecutionHandler{client: client}
}

// GetExecutions 获取执行列表
// @Summary 获取执行列表
// @Description 获取项目下的执行/迭代列表，支持按产品ID筛选
// @Tags 执行
// @Accept json
// @Produce json
// @Param projectID query int false "项目ID"
// @Param productID query int false "产品ID"
// @Success 200 {object} models.Response{data=[]models.Execution}
// @Failure 500 {object} models.Response
// @Router /api/executions [get]
func (h *ExecutionHandler) GetExecutions(c *gin.Context) {
	var executions []models.Execution

	projectIDStr := c.Query("projectID")
	productIDStr := c.Query("productID")

	if projectIDStr != "" {
		// 按项目筛选
		projectID, err := strconv.Atoi(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "无效的项目ID",
			})
			return
		}

		sdkExecutions, err := h.client.GetExecutions(projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "获取执行列表失败: " + err.Error(),
			})
			return
		}

		for _, e := range sdkExecutions {
			executions = append(executions, models.Execution{
				ID:      e.ID,
				Project: e.Project,
				Name:    e.Name,
				Code:    e.Code,
				Status:  e.Status,
				Type:    e.Type,
				Begin:   e.Begin,
				End:     e.End,
			})
		}
	} else if productIDStr != "" {
		// 按产品筛选（获取产品下所有项目的执行/迭代）
		productID, err := strconv.Atoi(productIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "无效的产品ID",
			})
			return
		}

		// 获取产品下的所有项目
		projects, err := h.client.GetProjectsByProduct(productID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "获取项目列表失败: " + err.Error(),
			})
			return
		}

		// 获取所有项目的执行/迭代
		for _, project := range projects {
			sdkExecutions, err := h.client.GetExecutions(project.ID)
			if err != nil {
				continue // 跳过获取失败的项目
			}

			for _, e := range sdkExecutions {
				executions = append(executions, models.Execution{
					ID:      e.ID,
					Project: e.Project,
					Name:    e.Name,
					Code:    e.Code,
					Status:  e.Status,
					Type:    e.Type,
					Begin:   e.Begin,
					End:     e.End,
				})
			}
		}
	} else {
		// 如果没有项目ID和产品ID，返回空列表
		c.JSON(http.StatusOK, models.Response{
			Code:    200,
			Message: "success",
			Data:    executions,
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "success",
		Data:    executions,
	})
}
