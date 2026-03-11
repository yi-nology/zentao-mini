package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/models"
	"chandao-mini/backend/core/zentao"
)

// ProductHandler 产品处理器
type ProductHandler struct {
	client *zentao.Client
}

// NewProductHandler 创建产品处理器
func NewProductHandler(client *zentao.Client) *ProductHandler {
	return &ProductHandler{client: client}
}

// GetProducts 获取产品列表
// @Summary 获取产品列表
// @Description 获取禅道系统中的所有产品列表
// @Tags 产品
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=[]models.Product}
// @Failure 500 {object} models.Response
// @Router /api/products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	products, err := h.client.GetProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "获取产品列表失败: " + err.Error(),
		})
		return
	}

	// 转换为模型
	var result []models.Product
	for _, p := range products {
		result = append(result, models.Product{
			ID:     p.ID,
			Name:   p.Name,
			Code:   p.Code,
			Type:   p.Type,
			Status: p.Status,
			Desc:   p.Desc,
		})
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "success",
		Data:    result,
	})
}
