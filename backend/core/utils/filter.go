package utils

import (
	"strings"
)

// FilterFunc 过滤函数类型
type FilterFunc[T any] func(item T) bool

// Filter 过滤切片
func Filter[T any](slice []T, filterFunc FilterFunc[T]) []T {
	if len(slice) == 0 {
		return []T{}
	}

	result := make([]T, 0)
	for _, item := range slice {
		if filterFunc(item) {
			result = append(result, item)
		}
	}
	return result
}

// FilterByDateRange 按日期范围过滤
// dateField 是获取日期字段的函数
func FilterByDateRange[T any](slice []T, startDate, endDate string, getDateFunc func(T) string) []T {
	if startDate == "" && endDate == "" {
		return slice
	}

	return Filter(slice, func(item T) bool {
		date := getDateFunc(item)
		if date == "" {
			return false
		}

		// 提取日期部分（去除时间）
		if len(date) > 10 {
			date = date[:10]
		}

		if startDate != "" && date < startDate {
			return false
		}
		if endDate != "" && date > endDate {
			return false
		}

		return true
	})
}

// FilterBySpecificDate 按具体日期过滤
func FilterBySpecificDate[T any](slice []T, specificDate string, getDateFunc func(T) string) []T {
	if specificDate == "" {
		return slice
	}

	return Filter(slice, func(item T) bool {
		date := getDateFunc(item)
		return strings.HasPrefix(date, specificDate)
	})
}

// FilterByDateRangeOrSpecific 按日期范围或具体日期过滤
// 如果有具体日期，优先使用具体日期过滤
func FilterByDateRangeOrSpecific[T any](slice []T, startDate, endDate, specificDate string, getDateFunc func(T) string) []T {
	if specificDate != "" {
		return FilterBySpecificDate(slice, specificDate, getDateFunc)
	}
	return FilterByDateRange(slice, startDate, endDate, getDateFunc)
}

// ChainFilter 链式过滤器
type ChainFilter[T any] struct {
	data []T
}

// NewChainFilter 创建链式过滤器
func NewChainFilter[T any](data []T) *ChainFilter[T] {
	return &ChainFilter[T]{data: data}
}

// Filter 执行过滤
func (cf *ChainFilter[T]) Filter(filterFunc FilterFunc[T]) *ChainFilter[T] {
	cf.data = Filter(cf.data, filterFunc)
	return cf
}

// FilterByDate 按日期范围过滤
func (cf *ChainFilter[T]) FilterByDate(startDate, endDate string, getDateFunc func(T) string) *ChainFilter[T] {
	cf.data = FilterByDateRange(cf.data, startDate, endDate, getDateFunc)
	return cf
}

// Paginate 执行分页
func (cf *ChainFilter[T]) Paginate(page, pageSize int) *ChainFilter[T] {
	cf.data = PaginateSlice(cf.data, page, pageSize)
	return cf
}

// Result 获取结果
func (cf *ChainFilter[T]) Result() []T {
	return cf.data
}

// Count 获取数量
func (cf *ChainFilter[T]) Count() int {
	return len(cf.data)
}
