package zentao

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yi-nology/zentao-mini/backend/core/config"
	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/metrics"
)

func TestNormalizeBugUserRefs(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string // bugs[0].assignedTo 期望的解析形态
	}{
		{
			name: "字符串账号包装成对象",
			in:   `{"page":1,"total":1,"limit":20,"bugs":[{"id":1,"assignedTo":"admin"}]}`,
			want: `{"account":"admin"}`,
		},
		{
			name: "对象形式保持不变",
			in:   `{"bugs":[{"id":1,"assignedTo":{"id":2,"account":"dev1","realname":"张三"}}]}`,
			want: `{"id":2,"account":"dev1","realname":"张三"}`,
		},
		{
			name: "openedBy 字符串同样处理",
			in:   `{"bugs":[{"id":1,"openedBy":"qa1"}]}`,
			want: `{"account":"qa1"}`,
		},
		{
			name: "响应无 bugs 字段时原样保留",
			in:   `{"error":"not found"}`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := normalizeBugUserRefs([]byte(tt.in))

			var doc map[string]json.RawMessage
			if err := json.Unmarshal(out, &doc); err != nil {
				t.Fatalf("输出不是合法 JSON: %v", err)
			}
			if tt.want == "" {
				if _, exists := doc["bugs"]; exists {
					t.Fatalf("不应产生 bugs 字段, got: %s", out)
				}
				return
			}

			var bugs []struct {
				AssignedTo json.RawMessage `json:"assignedTo"`
				OpenedBy   json.RawMessage `json:"openedBy"`
			}
			if err := json.Unmarshal(doc["bugs"], &bugs); err != nil {
				t.Fatalf("解析 bugs 失败: %v", err)
			}
			ref := bugs[0].AssignedTo
			if len(ref) == 0 {
				ref = bugs[0].OpenedBy
			}
			var got map[string]interface{}
			if err := json.Unmarshal(ref, &got); err != nil {
				t.Fatalf("assignedTo 不是对象: %v, raw=%s", err, ref)
			}
			var want map[string]interface{}
			_ = json.Unmarshal([]byte(tt.want), &want)
			if len(want) != len(got) {
				t.Errorf("字段数不一致: want %v, got %v", want, got)
			}
			for k, v := range want {
				if got[k] != v {
					t.Errorf("字段 %q: want %v, got %v", k, v, got[k])
				}
			}
		})
	}
}

// 旧版禅道返回字符串 assignedTo 时，GetBugs 应回退到容错解析而不是整体失败
func TestGetBugsFallbackForLegacyStringAssignedTo(t *testing.T) {
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
	mux.HandleFunc("/api.php/v1/products/1/bugs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"page":1,"total":2,"limit":20,"bugs":[
			{"id":1,"title":"旧格式bug","status":"active","assignedTo":"admin","openedBy":"qa1"},
			{"id":2,"title":"另一个","status":"resolved","assignedTo":{"id":3,"account":"dev1","realname":"李四"}}
		]}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "tester", "password")
	defer GlobalCache.Clear()

	bugs, err := client.GetBugs(1, 1, 20)
	if err != nil {
		t.Fatalf("GetBugs 应容错成功, got error: %v", err)
	}
	if len(bugs) != 2 {
		t.Fatalf("期望 2 条 bug, got %d", len(bugs))
	}
	if bugs[0].AssignedTo.Account != "admin" {
		t.Errorf("字符串 assignedTo 应包装成 UserRef.Account=admin, got %q", bugs[0].AssignedTo.Account)
	}
	if bugs[1].AssignedTo.Account != "dev1" || bugs[1].AssignedTo.Realname != "李四" {
		t.Errorf("对象 assignedTo 应原样保留, got %+v", bugs[1].AssignedTo)
	}
}
