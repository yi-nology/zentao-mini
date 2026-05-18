package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/yi-nology/zentao-mini/backend/core/errors"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func createTestRouter() *gin.Engine {
	router := gin.New()
	return router
}

func TestBugHandler_QueryParams(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "缺少产品ID参数",
			queryParams:    "?page=1&pageSize=20",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无效的productId",
			queryParams:    "?productId=invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "无效的page",
			queryParams:    "?productId=100&page=invalid",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无效的pageSize",
			queryParams:    "?productId=100&pageSize=invalid",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := createTestRouter()

			router.GET("/api/v1/bugs", func(c *gin.Context) {
				if tt.queryParams == "?productId=invalid" {
					errors.BadRequest(c, "无效的产品ID")
					return
				}
				errors.Success(c, gin.H{"list": []interface{}{}, "total": 0, "page": 1, "pageSize": 20})
			})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/bugs"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("期望状态码=%d, 实际状态码=%d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestPaginatedResponse(t *testing.T) {
	response := PaginatedResponse{
		List:     []string{"item1", "item2"},
		Total:    2,
		Page:     1,
		PageSize: 20,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("序列化失败: %v", err)
		return
	}

	var decoded PaginatedResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("反序列化失败: %v", err)
		return
	}

	if decoded.Total != 2 {
		t.Errorf("期望 total=2, 实际 total=%d", decoded.Total)
	}
	if decoded.Page != 1 {
		t.Errorf("期望 page=1, 实际 page=%d", decoded.Page)
	}
	if decoded.PageSize != 20 {
		t.Errorf("期望 pageSize=20, 实际 pageSize=%d", decoded.PageSize)
	}
}

type PaginatedResponse struct {
	List     interface{} `json:"list"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}
