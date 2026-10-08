package mcp

import (
	"testing"

	myzentao "github.com/yi-nology/zentao-mini/backend/core/zentao"
)

// TestFilterActionsByDate 动作日期窗过滤（扁鹊批次二百三十四）：闭区间
// [dateFrom, dateTo]，日期取 YYYY-MM-DD 前缀，空日期行剔除（统计口径诚实——
// 无日期不计数）。
func TestFilterActionsByDate(t *testing.T) {
	actions := []myzentao.BugAction{
		{Action: "opened", Date: "2026-10-07 18:00:00"},
		{Action: "resolved", Date: "2026-10-08 09:00:00"},
		{Action: "closed", Date: "2026-10-08 17:30:00"},
		{Action: "assigned", Date: ""}, // 无日期：剔除
	}
	filtered := filterActionsByDate(actions, "2026-10-08", "2026-10-08")
	if len(filtered) != 2 {
		t.Fatalf("当日动作应 2 条: %d", len(filtered))
	}
	if filtered[0].Action != "resolved" || filtered[1].Action != "closed" {
		t.Fatalf("过滤结果不符: %+v", filtered)
	}
	if got := countActionsByKind(filtered); got["resolved"] != 1 || got["closed"] != 1 {
		t.Fatalf("分组计数不符: %v", got)
	}
	// 空端=不约束
	if len(filterActionsByDate(actions, "", "")) != 4 {
		t.Fatal("空端应不约束")
	}
}

// TestGetBugActionsToolRegistered 工具面注册：get_bug_actions 只读（不进 writeTools，
// init 自动 readOnly 注解）、schema 必填 bugId。
func TestGetBugActionsToolRegistered(t *testing.T) {
	tl := GetToolByName("get_bug_actions")
	if tl == nil {
		t.Fatal("get_bug_actions 应已注册")
	}
	if tl.Annotations == nil || tl.Annotations.ReadOnlyHint == nil || !*tl.Annotations.ReadOnlyHint {
		t.Fatalf("get_bug_actions 应为只读注解: %+v", tl.Annotations)
	}
	if writeTools["get_bug_actions"] {
		t.Fatal("get_bug_actions 不得进写工具名单")
	}
	if len(tl.InputSchema.Required) == 0 || tl.InputSchema.Required[0] != "bugId" {
		t.Fatalf("bugId 应必填: %v", tl.InputSchema.Required)
	}
}
