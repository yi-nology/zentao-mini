package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"chandao-mini/backend/core/errors"
)

const (
	// DefaultPage 默认页码
	DefaultPage = 1
	// DefaultLimit 默认每页数量
	DefaultLimit = 20
	// MaxLimit 最大每页数量
	MaxLimit = 100
)

// PaginationParams 分页参数结构
type PaginationParams struct {
	Page  int `json:"page"`  // 当前页码
	Limit int `json:"limit"` // 每页数量
}

// ParsePagination 从gin.Context解析分页参数
// 支持 page/pageSize 和 page/limit 两种参数命名方式
func ParsePagination(c *gin.Context) PaginationParams {
	pageStr := c.DefaultQuery("page", strconv.Itoa(DefaultPage))

	// 优先使用pageSize，如果没有则使用limit
	limitStr := c.DefaultQuery("pageSize", c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)))

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = DefaultPage
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > MaxLimit {
		limit = DefaultLimit
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

// ParsePaginationWithMax 从gin.Context解析分页参数，允许自定义最大限制
func ParsePaginationWithMax(c *gin.Context, maxLimit int) PaginationParams {
	params := ParsePagination(c)
	if params.Limit > maxLimit {
		params.Limit = maxLimit
	}
	return params
}

// Paginate 执行分页计算
// 返回起始索引、结束索引
func Paginate(total, page, limit int) (start, end int) {
	if total == 0 {
		return 0, 0
	}

	start = (page - 1) * limit
	if start >= total {
		return 0, 0 // 超出范围
	}

	end = start + limit
	if end > total {
		end = total
	}

	return start, end
}

// PaginateSlice 对切片进行分页
// 返回分页后的切片，如果超出范围返回空切片
func PaginateSlice[T any](slice []T, page, limit int) []T {
	total := len(slice)
	start, end := Paginate(total, page, limit)

	if start >= total {
		return []T{}
	}

	return slice[start:end]
}

// PaginationMiddleware 分页中间件
// 自动解析分页参数并注入到context中
func PaginationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		params := ParsePagination(c)
		c.Set("pagination", params)
		c.Next()
	}
}

// GetPagination 从context获取分页参数
func GetPagination(c *gin.Context) PaginationParams {
	if params, exists := c.Get("pagination"); exists {
		return params.(PaginationParams)
	}
	return PaginationParams{
		Page:  DefaultPage,
		Limit: DefaultLimit,
	}
}

// ParseIntParam 解析整数参数
// 如果参数不存在或解析失败，返回错误
func ParseIntParam(c *gin.Context, paramName string) (int, error) {
	valueStr := c.Query(paramName)
	if valueStr == "" {
		return 0, errors.NewMissingParam(paramName)
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, errors.NewInvalidParam(paramName)
	}

	return value, nil
}

// ParseOptionalIntParam 解析可选的整数参数
// 如果参数不存在返回0和nil，如果解析失败返回错误
func ParseOptionalIntParam(c *gin.Context, paramName string) (int, error) {
	valueStr := c.Query(paramName)
	if valueStr == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, errors.NewInvalidParam(paramName)
	}

	return value, nil
}

// ParseRequiredIntParam 解析必需的整数参数
// 如果参数不存在或解析失败，返回错误
func ParseRequiredIntParam(c *gin.Context, paramName string, displayName string) (int, error) {
	valueStr := c.Query(paramName)
	if valueStr == "" {
		return 0, errors.NewMissingParam(paramName)
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, errors.New(errors.CodeInvalidID, "无效的"+displayName)
	}

	return value, nil
}
