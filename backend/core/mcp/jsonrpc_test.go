package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func newTestServer() *MCPServer {
	return NewMCPServerFromServices(nil, nil, nil, nil, nil, nil, nil, nil)
}

// 带正确 _meta 版本的请求包装
func metaReq(id, method string) []byte {
	return []byte(`{"jsonrpc":"2.0","id":"` + id + `","method":"` + method + `","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"` + ProtocolVersion + `"}}}`)
}

func TestJSONRPC_ServerDiscover(t *testing.T) {
	s := newTestServer()
	// server/discover 是唯一免版本检查的方法
	resp, fallback, _ := s.HandleJSONRPC([]byte(`{"jsonrpc":"2.0","id":7,"method":"server/discover"}`), nil)
	if fallback || resp == nil || resp.Error != nil {
		t.Fatalf("server/discover 应可用: %v", resp)
	}
	result := resp.Result.(map[string]interface{})
	versions := result["protocolVersions"].([]string)
	if len(versions) != 1 || versions[0] != ProtocolVersion {
		t.Errorf("protocolVersions = %v, want [%s]", versions, ProtocolVersion)
	}
	if result["latestVersion"] != ProtocolVersion {
		t.Errorf("latestVersion = %v", result["latestVersion"])
	}
	info := result["serverInfo"].(map[string]interface{})
	if info["name"] != "zentao-mini" {
		t.Errorf("serverInfo.name = %v", info["name"])
	}
}

func TestJSONRPC_MissingVersionRejected(t *testing.T) {
	s := newTestServer()
	// 2026-07-28 无握手：不带 _meta 版本的请求直接拒绝
	resp, _, _ := s.HandleJSONRPC([]byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`), nil)
	if resp == nil || resp.Error == nil || resp.Error.Code != codeUnsupportedProtocolVersion {
		t.Fatalf("缺版本应返回 -32022, got %+v", resp)
	}
	if !strings.Contains(resp.Error.Message, "missing") {
		t.Errorf("message 应说明缺失版本, got: %s", resp.Error.Message)
	}
	if resp.Error.Data == nil {
		t.Error("data 应携带正确版本指引")
	}
}

func TestJSONRPC_WrongVersionRejected(t *testing.T) {
	s := newTestServer()
	// 旧版本（如 2024-11-05 握手语义）一律拒绝
	resp, _, _ := s.HandleJSONRPC(
		[]byte(`{"jsonrpc":"2.0","id":2,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`), nil)
	if resp == nil || resp.Error == nil || resp.Error.Code != codeUnsupportedProtocolVersion {
		t.Fatalf("旧 initialize 应返回 -32022, got %+v", resp)
	}

	resp, _, _ = s.HandleJSONRPC(
		[]byte(`{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2025-06-18"}}}`), nil)
	if resp == nil || resp.Error == nil || resp.Error.Code != codeUnsupportedProtocolVersion {
		t.Fatalf("旧版本 tools/list 应返回 -32022, got %+v", resp)
	}
}

func TestJSONRPC_PingRemoved(t *testing.T) {
	s := newTestServer()
	resp, _, _ := s.HandleJSONRPC(metaReq("4", "ping"), nil)
	if resp == nil || resp.Error == nil || resp.Error.Code != codeMethodNotFound {
		t.Fatalf("ping 已从 2026-07-28 移除, 应返回 method not found, got %+v", resp)
	}
}

func TestJSONRPC_ToolsList(t *testing.T) {
	s := newTestServer()
	resp, fallback, _ := s.HandleJSONRPC(metaReq("2", "tools/list"), nil)
	if fallback || resp == nil || resp.Error != nil {
		t.Fatalf("tools/list 失败: %v", resp)
	}
	result := resp.Result.(map[string]interface{})
	tools := result["tools"].([]Tool)
	if len(tools) != len(Tools) {
		t.Errorf("tools 数量 = %d, want %d", len(tools), len(Tools))
	}
	if result["resultType"] != "complete" {
		t.Errorf("应带 resultType=complete, got %v", result["resultType"])
	}
	if result["ttlMs"] != toolsListTTLMS {
		t.Errorf("应带 ttlMs=%d, got %v", toolsListTTLMS, result["ttlMs"])
	}
}

func TestJSONRPC_ToolsCall_UnknownTool(t *testing.T) {
	s := newTestServer()
	resp, _, _ := s.HandleJSONRPC(
		[]byte(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"`+ProtocolVersion+`"},"name":"nope","arguments":{}}}`),
		nil,
	)
	if resp == nil || resp.Error == nil {
		t.Fatal("未知工具应返回 error")
	}
	if !strings.Contains(resp.Error.Message, "unknown tool") {
		t.Errorf("message = %s", resp.Error.Message)
	}
}

func TestJSONRPC_StructuredContent(t *testing.T) {
	s := newTestServer()
	resp, _, _ := s.HandleJSONRPC(
		[]byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"`+ProtocolVersion+`"},"name":"ping","arguments":{}}}`),
		nil,
	)
	if resp == nil || resp.Error != nil {
		t.Fatalf("tools/call ping 失败: %v", resp)
	}
	result := resp.Result.(map[string]interface{})
	if _, ok := result["structuredContent"]; !ok {
		t.Errorf("对象结果应附带 structuredContent, keys=%v", result)
	}
	if result["resultType"] != "complete" {
		t.Errorf("应带 resultType=complete, got %v", result["resultType"])
	}
}

func TestJSONRPC_AccessCheckOnToolsCall(t *testing.T) {
	s := newTestServer()
	blocked := func(action string) (bool, string) {
		if action == "get_bugs" {
			return true, "read-only mode: blocked"
		}
		return false, ""
	}
	body := `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"` + ProtocolVersion + `"},"name":"get_bugs","arguments":{}}}`
	resp, _, _ := s.HandleJSONRPC([]byte(body), blocked)
	if resp == nil || resp.Error == nil {
		t.Fatal("被访问控制拦截的 tools/call 应返回 error")
	}
	if !strings.Contains(resp.Error.Message, "read-only") {
		t.Errorf("message = %s", resp.Error.Message)
	}
}

func TestJSONRPC_LegacyActionFallsBack(t *testing.T) {
	s := newTestServer()
	resp, fallback, _ := s.HandleJSONRPC([]byte(`{"action":"ping"}`), nil)
	if !fallback || resp != nil {
		t.Errorf("zentao-mini 简化 action 协议应 fallback: resp=%v fallback=%v", resp, fallback)
	}
}

func TestJSONRPC_NotificationNoResponse(t *testing.T) {
	s := newTestServer()
	// 无 id 的通知一律不响应
	resp, _, _ := s.HandleJSONRPC([]byte(`{"jsonrpc":"2.0","method":"notifications/anything"}`), nil)
	if resp != nil {
		t.Errorf("通知不应产生响应, got %v", resp)
	}
}

func TestIsJSONRPCRequest(t *testing.T) {
	cases := []struct {
		body string
		want bool
	}{
		{`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`, true},
		{`{"method":"server/discover"}`, true},
		{`{"action":"ping"}`, false},
		{`not json`, false},
	}
	for _, c := range cases {
		if got := IsJSONRPCRequest([]byte(c.body)); got != c.want {
			t.Errorf("IsJSONRPCRequest(%s) = %v, want %v", c.body, got, c.want)
		}
	}
}

func TestJSONRPC_ResponseShape(t *testing.T) {
	s := newTestServer()
	resp, _, _ := s.HandleJSONRPC(metaReq("abc", "server/discover"), nil)
	out, _ := json.Marshal(resp)
	for _, key := range []string{`"jsonrpc":"2.0"`, `"id":"abc"`, `"result"`} {
		if !strings.Contains(string(out), key) {
			t.Errorf("响应缺少 %s: %s", key, out)
		}
	}
}
