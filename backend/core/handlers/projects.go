package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/models"
	"chandao-mini/backend/core/zentao"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	client *zentao.Client
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(client *zentao.Client) *ProjectHandler {
	return &ProjectHandler{client: client}
}

// GetProjects 获取项目列表
// @Summary 获取项目列表
// @Description 获取禅道系统中的所有项目列表，支持按产品筛选
// @Tags 项目
// @Accept json
// @Produce json
// @Param productID query int false "产品ID"
// @Success 200 {object} models.Response{data=[]models.Project}
// @Failure 500 {object} models.Response
// @Router /api/projects [get]
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	var projects []models.Project

	productIDStr := c.Query("productID")
	if productIDStr != "" {
		// 按产品筛选
		productID, err := strconv.Atoi(productIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "无效的产品ID",
			})
			return
		}

		sdkProjects, err := h.client.GetProjectsByProduct(productID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "获取项目列表失败: " + err.Error(),
			})
			return
		}

		for _, p := range sdkProjects {
			projects = append(projects, models.Project{
				ID:       p.ID,
				Name:     p.Name,
				Code:     p.Code,
				Model:    p.Model,
				Type:     p.Type,
				Status:   p.Status,
				PM:       p.PM,
				Begin:    p.Begin,
				End:      p.End,
				Progress: p.Progress,
			})
		}
	} else {
		// 获取所有项目
		sdkProjects, err := h.client.GetAllProjects(500)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Code:    500,
				Message: "获取项目列表失败: " + err.Error(),
			})
			return
		}

		for _, p := range sdkProjects {
			projects = append(projects, models.Project{
				ID:       p.ID,
				Name:     p.Name,
				Code:     p.Code,
				Model:    p.Model,
				Type:     p.Type,
				Status:   p.Status,
				PM:       p.PM,
				Begin:    p.Begin,
				End:      p.End,
				Progress: p.Progress,
			})
		}
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "success",
		Data:    projects,
	})
}
