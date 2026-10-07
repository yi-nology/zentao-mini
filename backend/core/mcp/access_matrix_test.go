package mcp

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/zentao-mini/backend/core/config"
)

// 分级权限模型矩阵：
//   读操作（get_bugs 等）  → 匿名可用（只要 MCP enabled）
//   写操作（add_bug_comment/transition_bug/create_bug...）
//     - 未配置 mcp.token          → 403（匿名只读，写能力未启用）
//     - 配置了 token + 提供正确 token → 放行
//     - 配置了 token + token 错误  → 401
//     - mcp.read_only=true        → 403（服务级只读，token 也不行）

func callCheckAccess(t *testing.T, method, path, action, bearer string) (bool, *app.RequestContext) {
	t.Helper()
	ctx := app.NewContext(0)
	ctx.Request.SetMethod(method)
	ctx.Request.SetRequestURI(path)
	if bearer != "" {
		ctx.Request.Header.Set("Authorization", "Bearer "+bearer)
	}
	blocked := checkAccess(ctx, action)
	return blocked, ctx
}

func initManager(enabled, readOnly bool, token string) {
	GetMCPModeManager().InitFromConfig(config.MCPConfig{
		Enabled:  enabled,
		ReadOnly: readOnly,
		Token:    token,
	})
}

func TestCheckAccess_ReadAnonymousAlways(t *testing.T) {
	// 未配置 token：读可用，写 403
	initManager(true, false, "")
	if blocked, _ := callCheckAccess(t, http.MethodGet, "/mcp/bugs", "get_bugs", ""); blocked {
		t.Error("未配置 token 时读操作应放行")
	}
	blocked, c := callCheckAccess(t, http.MethodPost, "/mcp", "transition_bug", "")
	if !blocked || c.Response.StatusCode() != http.StatusForbidden {
		t.Errorf("未配置 token 时写操作应 403, blocked=%v code=%d", blocked, c.Response.StatusCode())
	}
	if !strings.Contains(string(c.Response.Body()), "read-only") {
		t.Errorf("应提示匿名只读, got: %s", c.Response.Body())
	}

	// 配置 token：无凭证读仍可用
	initManager(true, false, "secret")
	if blocked, _ := callCheckAccess(t, http.MethodGet, "/mcp/bugs", "get_bugs", ""); blocked {
		t.Error("配置 token 后匿名读应放行（读不需要 token）")
	}
}

func TestCheckAccess_WriteTokenMatrix(t *testing.T) {
	initManager(true, false, "secret")

	// 正确 token → 写放行
	if blocked, _ := callCheckAccess(t, http.MethodPost, "/mcp", "transition_bug", "secret"); blocked {
		t.Error("正确 token 写操作应放行")
	}
	// 错误 token → 401
	blocked, c := callCheckAccess(t, http.MethodPost, "/mcp", "transition_bug", "wrong")
	if !blocked || c.Response.StatusCode() != http.StatusUnauthorized {
		t.Errorf("错误 token 写操作应 401, blocked=%v code=%d", blocked, c.Response.StatusCode())
	}
	// 缺 token → 401
	blocked, c = callCheckAccess(t, http.MethodPost, "/mcp", "transition_bug", "")
	if !blocked || c.Response.StatusCode() != http.StatusUnauthorized {
		t.Errorf("缺 token 写操作应 401, blocked=%v code=%d", blocked, c.Response.StatusCode())
	}

	// read_only=true：token 正确也 403（服务级只读）
	initManager(true, true, "secret")
	blocked, c = callCheckAccess(t, http.MethodPost, "/mcp", "transition_bug", "secret")
	if !blocked || c.Response.StatusCode() != http.StatusForbidden {
		t.Errorf("read_only 模式写应 403, blocked=%v code=%d", blocked, c.Response.StatusCode())
	}
}

func TestCheckAccess_MetaAndDisabled(t *testing.T) {
	initManager(true, false, "secret")

	// 元信息（action==""）免检查：tools/list 匿名可用
	if blocked, _ := callCheckAccess(t, http.MethodGet, "/mcp/tools", "", ""); blocked {
		t.Error("元信息端点应放行")
	}

	// 总开关关闭：全拒
	initManager(false, false, "")
	if blocked, c := callCheckAccess(t, http.MethodGet, "/mcp/bugs", "get_bugs", ""); !blocked || c.Response.StatusCode() != http.StatusServiceUnavailable {
		t.Errorf("disabled 应 503, code=%d", c.Response.StatusCode())
	}
}

// stdio：未配置 token 时写消息应在访问层被拒（读消息放行已由 TokenFromEnv 用例覆盖）
func TestStdioTransport_WriteBlockedWithoutToken(t *testing.T) {
	h := newStdioTestHarness()
	GetMCPModeManager().InitFromConfig(config.MCPConfig{Enabled: true})

	h.startListen()
	h.sendRequest(map[string]interface{}{"action": "transition_bug", "params": map[string]interface{}{"bugId": 1, "action": "close"}})
	h.sendRequest(map[string]interface{}{"action": "ping"})
	time.Sleep(50 * time.Millisecond)
	h.finish()

	resps := h.responses()
	if len(resps) < 2 {
		t.Fatalf("expected 2 responses, got %d: %v", len(resps), resps)
	}
	if resps[0]["status"] != "error" {
		t.Errorf("无 token 写应被拒, got %v", resps[0])
	}
	if !strings.Contains(resps[0]["message"].(string), "requires an MCP token") {
		t.Errorf("拒绝信息应说明需要 MCP token, got %v", resps[0]["message"])
	}
	if resps[1]["status"] != "ok" {
		t.Errorf("读消息（ping）应放行, got %v", resps[1])
	}
}
