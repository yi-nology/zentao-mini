package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/zentao"
)

// TimelogHandler 工时统计处理器
type TimelogHandler struct {
	client *zentao.Client
}

// NewTimelogHandler 创建工时统计处理器
func NewTimelogHandler(client *zentao.Client) *TimelogHandler {
	return &TimelogHandler{
		client: client,
	}
}

// GetTimelogAnalysis 获取工时统计分析
func (h *TimelogHandler) GetTimelogAnalysis(c *gin.Context) {
	// 获取查询参数
	productID := c.Query("productId")
	projectID := c.Query("projectId")
	executionID := c.Query("executionId")
	assignedTo := c.Query("assignedTo")
	dateFrom := c.Query("dateFrom")
	dateTo := c.Query("dateTo")

	// 验证必要参数
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "productId is required",
		})
		return
	}

	if dateFrom == "" || dateTo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "dateFrom and dateTo are required",
		})
		return
	}

	// 调用禅道客户端获取工时数据
	analysis, err := h.client.GetTimelogAnalysis(productID, projectID, executionID, assignedTo, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   analysis,
	})
}

// GetTimelogDashboard 获取工时统计看板数据
func (h *TimelogHandler) GetTimelogDashboard(c *gin.Context) {
	// 获取查询参数
	productID := c.Query("productId")
	projectID := c.Query("projectId")
	executionID := c.Query("executionId")
	assignedTo := c.Query("assignedTo")
	dateFrom := c.Query("dateFrom")
	dateTo := c.Query("dateTo")

	// 验证必要参数
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "productId is required",
		})
		return
	}

	if dateFrom == "" || dateTo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "dateFrom and dateTo are required",
		})
		return
	}

	// 调用禅道客户端获取工时数据
	analysis, err := h.client.GetTimelogAnalysis(productID, projectID, executionID, assignedTo, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	// 提取看板数据
	dashboardData := map[string]interface{}{
		"totalHours":  analysis["totalHours"],
		"effortCount": analysis["effortCount"],
		"taskCount":   analysis["taskCount"],
		"byProject":   analysis["byProject"],
		"byType":      analysis["byType"],
		"byDate":      analysis["byDate"],
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   dashboardData,
	})
}

// GetTimelogEfforts 获取工时流水明细
func (h *TimelogHandler) GetTimelogEfforts(c *gin.Context) {
	// 获取查询参数
	productID := c.Query("productId")
	projectID := c.Query("projectId")
	executionID := c.Query("executionId")
	assignedTo := c.Query("assignedTo")
	dateFrom := c.Query("dateFrom")
	dateTo := c.Query("dateTo")

	// 验证必要参数
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "productId is required",
		})
		return
	}

	if dateFrom == "" || dateTo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "dateFrom and dateTo are required",
		})
		return
	}

	// 调用禅道客户端获取工时数据
	analysis, err := h.client.GetTimelogAnalysis(productID, projectID, executionID, assignedTo, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	// 提取明细数据
	effortsData := analysis["efforts"]

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   effortsData,
	})
}
