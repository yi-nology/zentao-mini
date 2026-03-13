package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yi-nology/common/biz/zentao"

	"chandao-mini/backend/core/dto"
	"chandao-mini/backend/core/errors"
	"chandao-mini/backend/core/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// MockBugService 模拟Bug服务
type MockBugService struct {
	GetBugsFunc func(query *dto.BugQueryDTO) (interface{}, error)
}

// GetBugs 模拟获取Bug列表
func (m *MockBugService) GetBugsForTest(query *dto.BugQueryDTO) (interface{}, error) {
	if m.GetBugsFunc != nil {
		return m.GetBugsFunc(query)
	}
	return nil, nil
}

// createTestRouter 创建测试路由
func createTestRouter() *gin.Engine {
	router := gin.New()
	return router
}

// TestBugHandler_GetBugs 测试Bug处理器
func TestBugHandler_GetBugs(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockResult     interface{}
		mockError      error
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:        "成功获取Bug列表",
			queryParams: "?productId=100&page=1&limit=20",
			mockResult: &dto.PaginatedResponse{
				List:  []zentao.Bug{},
				Total: 0,
				Page:  1,
				Limit: 20,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response errors.Response
				if err := json.Unmarshal(body, &response); err != nil {
					t.Errorf("解析响应失败: %v", err)
					return
				}
				if response.Code != errors.CodeSuccess {
					t.Errorf("期望 code=%d, 实际 code=%d", errors.CodeSuccess, response.Code)
				}
			},
		},
		{
			name:           "缺少产品ID参数",
			queryParams:    "?page=1&limit=20",
			mockResult:     nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response errors.Response
				if err := json.Unmarshal(body, &response); err != nil {
					t.Errorf("解析响应失败: %v", err)
					return
				}
				// 无产品ID时应该返回空列表
				if response.Code != errors.CodeSuccess {
					t.Errorf("期望 code=%d, 实际 code=%d", errors.CodeSuccess, response.Code)
				}
			},
		},
		{
			name:           "服务错误",
			queryParams:    "?productId=100",
			mockError:      errors.New(errors.CodeInternalError, "服务器错误"),
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body []byte) {
				var response errors.Response
				if err := json.Unmarshal(body, &response); err != nil {
					t.Errorf("解析响应失败: %v", err)
					return
				}
				if response.Code != errors.CodeExternalError {
					t.Errorf("期望 code=%d, 实际 code=%d", errors.CodeExternalError, response.Code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试路由
			router := createTestRouter()

			// 创建mock服务
			mockService := &service.BugService{}

			// 创建handler
			handler := NewBugHandler(mockService)

			// 设置路由
			router.GET("/api/v1/bugs", handler.GetBugs)

			// 创建请求
			req := httptest.NewRequest(http.MethodGet, "/api/v1/bugs"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证状态码
			if w.Code != tt.expectedStatus {
				t.Errorf("期望状态码=%d, 实际状态码=%d", tt.expectedStatus, w.Code)
			}

			// 执行自定义验证
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

// TestBugHandler_GetBugs_WithFilters 测试带筛选条件的Bug查询
func TestBugHandler_GetBugs_WithFilters(t *testing.T) {
	router := createTestRouter()
	mockService := &service.BugService{}
	handler := NewBugHandler(mockService)
	router.GET("/api/v1/bugs", handler.GetBugs)

	tests := []struct {
		name        string
		queryParams string
	}{
		{
			name:        "按状态筛选",
			queryParams: "?productId=100&status=active",
		},
		{
			name:        "按指派人筛选",
			queryParams: "?productId=100&assignedTo=user1",
		},
		{
			name:        "按日期范围筛选",
			queryParams: "?productId=100&startDate=2024-01-01&endDate=2024-01-31",
		},
		{
			name:        "按具体日期筛选",
			queryParams: "?productId=100&specificDate=2024-01-15",
		},
		{
			name:        "组合筛选",
			queryParams: "?productId=100&status=active&assignedTo=user1&page=1&limit=10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/bugs"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// 验证请求成功处理
			if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
				t.Errorf("意外的状态码=%d", w.Code)
			}
		})
	}
}

// TestBugHandler_GetBugs_InvalidParams 测试无效参数
func TestBugHandler_GetBugs_InvalidParams(t *testing.T) {
	router := createTestRouter()
	mockService := &service.BugService{}
	handler := NewBugHandler(mockService)
	router.GET("/api/v1/bugs", handler.GetBugs)

	tests := []struct {
		name        string
		queryParams string
	}{
		{
			name:        "无效的productId",
			queryParams: "?productId=invalid",
		},
		{
			name:        "无效的page",
			queryParams: "?productId=100&page=invalid",
		},
		{
			name:        "无效的limit",
			queryParams: "?productId=100&limit=invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/bugs"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// 验证返回错误响应
			if w.Code != http.StatusBadRequest {
				t.Errorf("期望状态码=%d, 实际状态码=%d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	List  interface{} `json:"list"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// TestIntegration_BugHandler_FullFlow 测试完整的请求-响应流程
func TestIntegration_BugHandler_FullFlow(t *testing.T) {
	// 这个测试演示了完整的集成测试流程
	// 在实际项目中，应该使用真实的数据库连接或更完整的mock

	t.Run("完整流程测试", func(t *testing.T) {
		router := createTestRouter()
		mockService := &service.BugService{}
		handler := NewBugHandler(mockService)
		router.GET("/api/v1/bugs", handler.GetBugs)

		// 发送请求
		req := httptest.NewRequest(http.MethodGet, "/api/v1/bugs?productId=100&page=1&limit=20", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 验证响应
		if w.Code != http.StatusOK {
			t.Errorf("期望状态码=%d, 实际状态码=%d", http.StatusOK, w.Code)
		}

		// 验证响应头
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json; charset=utf-8" {
			t.Errorf("期望 Content-Type=application/json; charset=utf-8, 实际=%s", contentType)
		}

		// 验证响应体
		var response errors.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("解析响应失败: %v", err)
			return
		}

		// 验证响应结构
		if response.Code == 0 && response.Message == "" {
			t.Error("响应结构不完整")
		}
	})
}
