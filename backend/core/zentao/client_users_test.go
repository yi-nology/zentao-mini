package zentao

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yi-nology/zentao-mini/backend/core/config"
	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/metrics"
)

// newUsersPaginationServer 起 8 个用户的假禅道上游（token 模式 + v1 users 分页）
func newUsersPaginationServer(t *testing.T) *httptest.Server {
	t.Helper()
	if err := metrics.Init(); err != nil {
		t.Fatalf("metrics 初始化失败: %v", err)
	}
	if err := logger.Init(&config.LogConfig{Level: "error", Format: "console"}); err != nil {
		t.Fatalf("logger 初始化失败: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api.php/v1/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"test-token"}`))
	})
	mux.HandleFunc("/api.php/v1/users", func(w http.ResponseWriter, r *http.Request) {
		var page int
		fmt.Sscanf(r.URL.Query().Get("page"), "%d", &page)
		users := []map[string]interface{}{}
		if page == 1 {
			for i := range 8 {
				account := string(rune('a' + i))
				users = append(users, map[string]interface{}{
					"id": i + 1, "account": account, "realname": "用户" + account,
				})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"page": page, "total": 8, "limit": 100, "users": users,
		})
	})
	return httptest.NewServer(mux)
}

// TestGetUsers_PageOverflow issue #31：page 无上界时 (page-1)*pageSize 溢出为负，
// 负下标切片 panic（对外 500）。修复后越界页应稳定返回空列表而非 panic。
func TestGetUsers_PageOverflow(t *testing.T) {
	server := newUsersPaginationServer(t)
	defer server.Close()

	client := NewClient(server.URL, "tester", "password")
	defer GlobalCache.Clear()

	tests := []struct {
		name        string
		page        int
		pageSize    int
		expectedLen int
	}{
		{name: "第一页正常", page: 1, pageSize: 5, expectedLen: 5},
		{name: "尾页截断", page: 2, pageSize: 5, expectedLen: 3},
		{name: "越界页返回空", page: 3, pageSize: 5, expectedLen: 0},
		// page=2^62：(page-1)*20 溢出为 -20，修复前 allUsers[-20:0] panic
		{name: "page=2^62溢出为负下标", page: 4611686018427387904, pageSize: 20, expectedLen: 0},
		// page=MaxInt64：(page-1)*20 溢出为 -40，修复前 allUsers[-40:-20] panic
		{name: "page=MaxInt64溢出为负下标", page: 9223372036854775807, pageSize: 20, expectedLen: 0},
		// 2^61 量级：溢出回绕后 start 为正大数，走 start >= total 分支返回空
		{name: "page=2^61回绕为正大数", page: 2305843009213693952, pageSize: 20, expectedLen: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GlobalCache.Clear() // 每个用例拿干净缓存，避免上游数据串场
			resp, err := client.GetUsers(tt.page, tt.pageSize)
			if err != nil {
				t.Fatalf("GetUsers(%d, %d) err = %v", tt.page, tt.pageSize, err)
			}
			if resp.Total != 8 {
				t.Errorf("Total = %d, want 8", resp.Total)
			}
			if len(resp.Users) != tt.expectedLen {
				t.Errorf("len(Users) = %d, want %d (page=%d pageSize=%d)",
					len(resp.Users), tt.expectedLen, tt.page, tt.pageSize)
			}
		})
	}
}
