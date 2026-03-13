package handlers

import (
	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/service"
)

// ProductHandler 产品处理器
// 只负责HTTP请求/响应处理，业务逻辑由Service层处理
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler 创建产品处理器
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// GetProducts 获取产品列表
// @Summary 获取产品列表
// @Description 获取禅道系统中的所有产品列表
// @Tags 产品
// @Accept json
// @Produce json
// @Success 200 {object} errors.Response{data=[]vo.ProductVO}
// @Failure 500 {object} errors.Response
// @Router /api/v1/products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	// 调用Service层处理业务逻辑
	result, err := h.productService.GetProducts()
	if err != nil {
		errors.Error(c, errors.ExternalError("禅道", err))
		return
	}

	// 返回成功响应
	errors.Success(c, result)
}
