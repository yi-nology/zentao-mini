package zentao

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yi-nology/zentao-mini/backend/core/config"
	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/metrics"
)

// newWriteTestServer 起 mock 禅道并记录最近一次写请求
func newWriteTestServer(t *testing.T) (*httptest.Server, *capturedRequest) {
	t.Helper()
	if err := metrics.Init(); err != nil {
		t.Fatalf("metrics 初始化失败: %v", err)
	}
	if err := logger.Init(&config.LogConfig{Level: "error", Format: "console"}); err != nil {
		t.Fatalf("logger 初始化失败: %v", err)
	}
	captured := &capturedRequest{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api.php/v1/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"test-token"}`))
	})
	handlers := map[string]http.HandlerFunc{
		"/api.php/v1/bugs/123/confirm":  writeOK,
		"/api.php/v1/bugs/123/resolve":  writeOK,
		"/api.php/v1/bugs/123/close":    writeOK,
		"/api.php/v1/bugs/123/activate": writeOK,
		"/api.php/v1/bugs/123/assign":   writeOK,
	}
	for path, h := range handlers {
		p, handler := path, h
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			captured.path = r.URL.Path
			captured.method = r.Method
			_ = json.Unmarshal(body, &captured.body)
			handler(w, r)
		})
	}
	server := httptest.NewServer(mux)
	return server, captured
}

type capturedRequest struct {
	path   string
	method string
	body   map[string]interface{}
}

func writeOK(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"id":123,"status":"resolved"}`))
}

func TestBugWriteOperations(t *testing.T) {
	server, captured := newWriteTestServer(t)
	defer server.Close()

	client := NewClient(server.URL, "tester", "password")
	defer GlobalCache.Clear()

	if err := client.ConfirmBug(123, "确认一下"); err != nil {
		t.Fatalf("ConfirmBug 失败: %v", err)
	}
	if captured.path != "/api.php/v1/bugs/123/confirm" || captured.method != http.MethodPost {
		t.Errorf("confirm 请求路径/方法不对: %s %s", captured.method, captured.path)
	}
	if captured.body["comment"] != "确认一下" {
		t.Errorf("confirm 应携带 comment, got %v", captured.body)
	}

	if err := client.ResolveBug(123, "fixed", "trunk", "已修复"); err != nil {
		t.Fatalf("ResolveBug 失败: %v", err)
	}
	if captured.body["resolution"] != "fixed" || captured.body["resolvedBuild"] != "trunk" || captured.body["comment"] != "已修复" {
		t.Errorf("resolve 载荷不对: %v", captured.body)
	}

	if err := client.CloseBug(123, "验证通过"); err != nil {
		t.Fatalf("CloseBug 失败: %v", err)
	}
	if captured.body["comment"] != "验证通过" {
		t.Errorf("close 载荷不对: %v", captured.body)
	}

	if err := client.ActivateBug(123, "dev1", "回归复现"); err != nil {
		t.Fatalf("ActivateBug 失败: %v", err)
	}
	if captured.body["assignedTo"] != "dev1" {
		t.Errorf("activate 载荷不对: %v", captured.body)
	}

	if err := client.AssignBug(123, "qa1", "麻烦复测"); err != nil {
		t.Fatalf("AssignBug 失败: %v", err)
	}
	if captured.body["assignedTo"] != "qa1" {
		t.Errorf("assign 载荷不对: %v", captured.body)
	}
}
