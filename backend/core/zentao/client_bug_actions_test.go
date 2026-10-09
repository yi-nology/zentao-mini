package zentao

import (
	"strings"
	"testing"
)

// TestParseBugActionsResponse 动作历史响应解析容错（扁鹊批次二百三十四）：禅道
// api-getModel 的返回形态随版本/状态漂移——关联数组（惯例）/数组/空历史/失败/登录页
// HTML 五形态必须各得其所，错误原样透出不静默。
func TestParseBugActionsResponse(t *testing.T) {
	// ① 关联数组形态（api-getModel 惯例：data 为 id→action 对象）
	actions, err := parseBugActionsResponse(`{"status":"success","data":{"3":{"id":3,"actor":"zhangsan","action":"opened","date":"2026-10-08 09:12:00","comment":""},"7":{"id":7,"actor":"lisi","action":"resolved","date":"2026-10-08 15:40:30","comment":"修复于 v2.3","extra":"fixed"}}}`)
	if err != nil {
		t.Fatalf("关联数组形态应解析成功: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("应解析出 2 条动作: %d", len(actions))
	}

	// ② 数组形态
	actions, err = parseBugActionsResponse(`{"status":"success","data":[{"id":3,"actor":"zhangsan","action":"opened","date":"2026-10-08 09:12:00"}]}`)
	if err != nil {
		t.Fatalf("数组形态应解析成功: %v", err)
	}
	if len(actions) != 1 || actions[0].Action != "opened" {
		t.Fatalf("数组形态解析不符: %+v", actions)
	}

	// ③ 空历史（data=[] / null / 缺 data）
	for _, body := range []string{
		`{"status":"success","data":[]}`,
		`{"status":"success","data":null}`,
		`{"status":"success"}`,
	} {
		actions, err = parseBugActionsResponse(body)
		if err != nil || len(actions) != 0 {
			t.Fatalf("空历史应零错误零记录: %v, %v", actions, err)
		}
	}

	// ④ 禅道侧失败：错误原样透出
	_, err = parseBugActionsResponse(`{"status":"failed","message":"no such method"}`)
	if err == nil || !strings.Contains(err.Error(), "no such method") {
		t.Fatalf("失败形态应透出禅道错误: %v", err)
	}

	// ⑤ 登录页 HTML（会话失效）：报「非 JSON」供上层换形态/重登
	_, err = parseBugActionsResponse(`<!DOCTYPE html><html><body>user-login</body></html>`)
	if err == nil || !strings.Contains(err.Error(), "非 JSON") {
		t.Fatalf("HTML 形态应报非 JSON（未登录）: %v", err)
	}

	// ⑥ data 是 JSON 字符串且内嵌 user-deny-api-getmodel（v1.6.0 实弹 pm.kylin.com 形态，
	// ZenTao 安全开关拒 api-getModel）：必须显式报错带修复路径，绝不静默当空表。
	_, err = parseBugActionsResponse(`{"status":"success","data":"{\"locate\":\"https://pm.example.com/user-deny-api-getmodel.json\"}","md5":"x"}`)
	if err == nil || !strings.Contains(err.Error(), "api-getModel") || !strings.Contains(err.Error(), "禁用") {
		t.Fatalf("deny 形态应显式报错带修复路径: %v", err)
	}

	// ⑦ data 是其他重定向字符串：显式报错不吞。
	_, err = parseBugActionsResponse(`{"status":"success","data":"{\"locate\":\"https://pm.example.com/elsewhere\"}"}`)
	if err == nil || !strings.Contains(err.Error(), "重定向") {
		t.Fatalf("重定向形态应显式报错: %v", err)
	}

	// ⑧ data 是不可识别对象（无动作键值）：显式报错不吞（v1.6.0 静默空表根因收口）。
	_, err = parseBugActionsResponse(`{"status":"success","data":{"title":"bug","steps":"x"}}`)
	if err == nil || !strings.Contains(err.Error(), "形态未识别") {
		t.Fatalf("未识别对象形态应显式报错: %v", err)
	}
}
