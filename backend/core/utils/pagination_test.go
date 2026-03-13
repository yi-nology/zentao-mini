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

// TestParsePagination 测试分页参数解析
func TestParsePagination(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedPage   int
		expectedLimit  int
	}{
		{
			name:          "默认值",
			queryParams:   map[string]string{},
			expectedPage:  DefaultPage,
			expectedLimit: DefaultLimit,
		},
		{
			name: "自定义page和limit",
			queryParams: map[string]string{
				"page":  "2",
				"limit": "50",
			},
			expectedPage:  2,
			expectedLimit: 50,
		},
		{
			name: "使用pageSize参数",
			queryParams: map[string]string{
				"page":     "3",
				"pageSize": "30",
			},
			expectedPage:  3,
			expectedLimit: 30,
		},
		{
			name: "page参数无效",
			queryParams: map[string]string{
				"page":  "invalid",
				"limit": "20",
			},
			expectedPage:  DefaultPage,
			expectedLimit: 20,
		},
		{
			name: "limit参数无效",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "invalid",
			},
			expectedPage:  1,
			expectedLimit: DefaultLimit,
		},
		{
			name: "limit超过最大值",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "200",
			},
			expectedPage:  1,
			expectedLimit: DefaultLimit,
		},
		{
			name: "limit小于1",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "0",
			},
			expectedPage:  1,
			expectedLimit: DefaultLimit,
		},
		{
			name: "page小于1",
			queryParams: map[string]string{
				"page":  "0",
				"limit": "20",
			},
			expectedPage:  DefaultPage,
			expectedLimit: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试HTTP请求
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			c.Request = req
			
			// 设置查询参数
			q := c.Request.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			c.Request.URL.RawQuery = q.Encode()

			// 执行测试
			params := ParsePagination(c)

			// 验证结果
			if params.Page != tt.expectedPage {
				t.Errorf("期望 page=%d, 实际 page=%d", tt.expectedPage, params.Page)
			}
			if params.Limit != tt.expectedLimit {
				t.Errorf("期望 limit=%d, 实际 limit=%d", tt.expectedLimit, params.Limit)
			}
		})
	}
}

// TestParsePaginationWithMax 测试带最大限制的分页参数解析
func TestParsePaginationWithMax(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		maxLimit      int
		expectedLimit int
	}{
		{
			name: "limit小于maxLimit",
			queryParams: map[string]string{
				"limit": "50",
			},
			maxLimit:      100,
			expectedLimit: 50,
		},
		{
			name: "limit大于maxLimit",
			queryParams: map[string]string{
				"limit": "80",
			},
			maxLimit:      50,
			expectedLimit: 50,
		},
		{
			name: "limit等于maxLimit",
			queryParams: map[string]string{
				"limit": "100",
			},
			maxLimit:      100,
			expectedLimit: 100,
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

			params := ParsePaginationWithMax(c, tt.maxLimit)

			if params.Limit != tt.expectedLimit {
				t.Errorf("期望 limit=%d, 实际 limit=%d", tt.expectedLimit, params.Limit)
			}
		})
	}
}

// TestPaginate 测试分页计算
func TestPaginate(t *testing.T) {
	tests := []struct {
		name         string
		total        int
		page         int
		limit        int
		expectedStart int
		expectedEnd   int
	}{
		{
			name:          "第一页",
			total:         100,
			page:          1,
			limit:         20,
			expectedStart: 0,
			expectedEnd:   20,
		},
		{
			name:          "中间页",
			total:         100,
			page:          3,
			limit:         20,
			expectedStart: 40,
			expectedEnd:   60,
		},
		{
			name:          "最后一页",
			total:         100,
			page:          5,
			limit:         20,
			expectedStart: 80,
			expectedEnd:   100,
		},
		{
			name:          "超出范围",
			total:         100,
			page:          10,
			limit:         20,
			expectedStart: 0,
			expectedEnd:   0,
		},
		{
			name:          "空列表",
			total:         0,
			page:          1,
			limit:         20,
			expectedStart: 0,
			expectedEnd:   0,
		},
		{
			name:          "部分页",
			total:         95,
			page:          5,
			limit:         20,
			expectedStart: 80,
			expectedEnd:   95,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := Paginate(tt.total, tt.page, tt.limit)
			
			if start != tt.expectedStart {
				t.Errorf("期望 start=%d, 实际 start=%d", tt.expectedStart, start)
			}
			if end != tt.expectedEnd {
				t.Errorf("期望 end=%d, 实际 end=%d", tt.expectedEnd, end)
			}
		})
	}
}

// TestPaginateSlice 测试切片分页
func TestPaginateSlice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		page     int
		limit    int
		expected []int
	}{
		{
			name:     "第一页",
			slice:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			page:     1,
			limit:    3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "中间页",
			slice:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			page:     2,
			limit:    3,
			expected: []int{4, 5, 6},
		},
		{
			name:     "最后一页",
			slice:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			page:     4,
			limit:    3,
			expected: []int{10},
		},
		{
			name:     "超出范围",
			slice:    []int{1, 2, 3},
			page:     10,
			limit:    3,
			expected: []int{},
		},
		{
			name:     "空切片",
			slice:    []int{},
			page:     1,
			limit:    3,
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PaginateSlice(tt.slice, tt.page, tt.limit)
			
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

// TestPaginationMiddleware 测试分页中间件
func TestPaginationMiddleware(t *testing.T) {
	// 设置测试路由
	router := gin.New()
	router.Use(PaginationMiddleware())
	router.GET("/test", func(c *gin.Context) {
		params := GetPagination(c)
		c.JSON(http.StatusOK, params)
	})

	// 测试默认值
	t.Run("默认值", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		// 验证响应状态码
		if w.Code != http.StatusOK {
			t.Errorf("期望状态码=%d, 实际状态码=%d", http.StatusOK, w.Code)
		}
	})

	// 测试自定义参数
	t.Run("自定义参数", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test?page=2&limit=50", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码=%d, 实际状态码=%d", http.StatusOK, w.Code)
		}
	})
}

// TestGetPagination 测试从context获取分页参数
func TestGetPagination(t *testing.T) {
	t.Run("从context获取", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		params := PaginationParams{
			Page:  3,
			Limit: 30,
		}
		c.Set("pagination", params)
		
		result := GetPagination(c)
		
		if result.Page != params.Page {
			t.Errorf("期望 page=%d, 实际 page=%d", params.Page, result.Page)
		}
		if result.Limit != params.Limit {
			t.Errorf("期望 limit=%d, 实际 limit=%d", params.Limit, result.Limit)
		}
	})

	t.Run("context中没有值", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		result := GetPagination(c)
		
		if result.Page != DefaultPage {
			t.Errorf("期望 page=%d, 实际 page=%d", DefaultPage, result.Page)
		}
		if result.Limit != DefaultLimit {
			t.Errorf("期望 limit=%d, 实际 limit=%d", DefaultLimit, result.Limit)
		}
	})
}
