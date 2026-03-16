package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name             string
		queryParams      map[string]string
		expectedPage     int
		expectedPageSize int
	}{
		{
			name:             "默认值",
			queryParams:      map[string]string{},
			expectedPage:     DefaultPage,
			expectedPageSize: DefaultPageSize,
		},
		{
			name: "自定义page和pageSize",
			queryParams: map[string]string{
				"page":     "2",
				"pageSize": "50",
			},
			expectedPage:     2,
			expectedPageSize: 50,
		},
		{
			name: "page参数无效",
			queryParams: map[string]string{
				"page":     "invalid",
				"pageSize": "20",
			},
			expectedPage:     DefaultPage,
			expectedPageSize: 20,
		},
		{
			name: "pageSize参数无效",
			queryParams: map[string]string{
				"page":     "1",
				"pageSize": "invalid",
			},
			expectedPage:     1,
			expectedPageSize: DefaultPageSize,
		},
		{
			name: "pageSize超过最大值",
			queryParams: map[string]string{
				"page":     "1",
				"pageSize": "200",
			},
			expectedPage:     1,
			expectedPageSize: DefaultPageSize,
		},
		{
			name: "pageSize小于1",
			queryParams: map[string]string{
				"page":     "1",
				"pageSize": "0",
			},
			expectedPage:     1,
			expectedPageSize: DefaultPageSize,
		},
		{
			name: "page小于1",
			queryParams: map[string]string{
				"page":     "0",
				"pageSize": "20",
			},
			expectedPage:     DefaultPage,
			expectedPageSize: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			c.Request = req

			q := c.Request.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			c.Request.URL.RawQuery = q.Encode()

			params := ParsePagination(c)

			if params.Page != tt.expectedPage {
				t.Errorf("期望 page=%d, 实际 page=%d", tt.expectedPage, params.Page)
			}
			if params.PageSize != tt.expectedPageSize {
				t.Errorf("期望 pageSize=%d, 实际 pageSize=%d", tt.expectedPageSize, params.PageSize)
			}
		})
	}
}

func TestParsePaginationWithMax(t *testing.T) {
	tests := []struct {
		name             string
		queryParams      map[string]string
		maxPageSize      int
		expectedPageSize int
	}{
		{
			name: "pageSize小于maxPageSize",
			queryParams: map[string]string{
				"pageSize": "50",
			},
			maxPageSize:      100,
			expectedPageSize: 50,
		},
		{
			name: "pageSize大于maxPageSize",
			queryParams: map[string]string{
				"pageSize": "80",
			},
			maxPageSize:      50,
			expectedPageSize: 50,
		},
		{
			name: "pageSize等于maxPageSize",
			queryParams: map[string]string{
				"pageSize": "100",
			},
			maxPageSize:      100,
			expectedPageSize: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			c.Request = req

			q := c.Request.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			c.Request.URL.RawQuery = q.Encode()

			params := ParsePaginationWithMax(c, tt.maxPageSize)

			if params.PageSize != tt.expectedPageSize {
				t.Errorf("期望 pageSize=%d, 实际 pageSize=%d", tt.expectedPageSize, params.PageSize)
			}
		})
	}
}

func TestPaginate(t *testing.T) {
	tests := []struct {
		name          string
		total         int
		page          int
		pageSize      int
		expectedStart int
		expectedEnd   int
	}{
		{
			name:          "第一页",
			total:         100,
			page:          1,
			pageSize:      20,
			expectedStart: 0,
			expectedEnd:   20,
		},
		{
			name:          "中间页",
			total:         100,
			page:          3,
			pageSize:      20,
			expectedStart: 40,
			expectedEnd:   60,
		},
		{
			name:          "最后一页",
			total:         100,
			page:          5,
			pageSize:      20,
			expectedStart: 80,
			expectedEnd:   100,
		},
		{
			name:          "超出范围",
			total:         100,
			page:          10,
			pageSize:      20,
			expectedStart: 0,
			expectedEnd:   0,
		},
		{
			name:          "空列表",
			total:         0,
			page:          1,
			pageSize:      20,
			expectedStart: 0,
			expectedEnd:   0,
		},
		{
			name:          "部分页",
			total:         95,
			page:          5,
			pageSize:      20,
			expectedStart: 80,
			expectedEnd:   95,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := Paginate(tt.total, tt.page, tt.pageSize)

			if start != tt.expectedStart {
				t.Errorf("期望 start=%d, 实际 start=%d", tt.expectedStart, start)
			}
			if end != tt.expectedEnd {
				t.Errorf("期望 end=%d, 实际 end=%d", tt.expectedEnd, end)
			}
		})
	}
}

func TestPaginateSlice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		page     int
		pageSize int
		expected []int
	}{
		{
			name:     "第一页",
			slice:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			page:     1,
			pageSize: 3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "中间页",
			slice:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			page:     2,
			pageSize: 3,
			expected: []int{4, 5, 6},
		},
		{
			name:     "最后一页",
			slice:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			page:     4,
			pageSize: 3,
			expected: []int{10},
		},
		{
			name:     "超出范围",
			slice:    []int{1, 2, 3},
			page:     10,
			pageSize: 3,
			expected: []int{},
		},
		{
			name:     "空切片",
			slice:    []int{},
			page:     1,
			pageSize: 3,
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PaginateSlice(tt.slice, tt.page, tt.pageSize)

			if len(result) != len(tt.expected) {
				t.Errorf("期望长度=%d, 实际长度=%d", len(tt.expected), len(result))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("索引 %d: 期望=%d, 实际=%d", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestPaginationMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(PaginationMiddleware())
	router.GET("/test", func(c *gin.Context) {
		params := GetPagination(c)
		c.JSON(http.StatusOK, params)
	})

	t.Run("默认值", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码=%d, 实际状态码=%d", http.StatusOK, w.Code)
		}
	})

	t.Run("自定义参数", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test?page=2&pageSize=50", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码=%d, 实际状态码=%d", http.StatusOK, w.Code)
		}
	})
}

func TestGetPagination(t *testing.T) {
	t.Run("从context获取", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		params := PaginationParams{
			Page:     3,
			PageSize: 30,
		}
		c.Set("pagination", params)

		result := GetPagination(c)

		if result.Page != params.Page {
			t.Errorf("期望 page=%d, 实际 page=%d", params.Page, result.Page)
		}
		if result.PageSize != params.PageSize {
			t.Errorf("期望 pageSize=%d, 实际 pageSize=%d", params.PageSize, result.PageSize)
		}
	})

	t.Run("context中没有值", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		result := GetPagination(c)

		if result.Page != DefaultPage {
			t.Errorf("期望 page=%d, 实际 page=%d", DefaultPage, result.Page)
		}
		if result.PageSize != DefaultPageSize {
			t.Errorf("期望 pageSize=%d, 实际 pageSize=%d", DefaultPageSize, result.PageSize)
		}
	})
}
