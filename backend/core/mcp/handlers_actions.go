package mcp

import (
	"strings"

	myzentao "github.com/yi-nology/zentao-mini/backend/core/zentao"
)

// handleGetBugActions 获取 Bug 动作历史（流转/处置记录，只读）——「今天流转多少」
// 统计与处置审计的承接面（扁鹊批次二百三十四）。dateFrom/dateTo（YYYY-MM-DD，可选）
// 按动作日期过滤；响应带 total，按 action 分组计数随行（opened/resolved/closed/
// assigned/commented 等单 key 计数，流转统计免二次加工）。
func (s *MCPServer) handleGetBugActions(params map[string]interface{}) (interface{}, error) {
	bugID, err := bugIDFromParams(params)
	if err != nil {
		return nil, err
	}
	dateFrom := strings.TrimSpace(stringFromParams(params, "dateFrom"))
	dateTo := strings.TrimSpace(stringFromParams(params, "dateTo"))

	actions, err := s.bugService.GetBugActions(bugID)
	if err != nil {
		return nil, err
	}
	if dateFrom != "" || dateTo != "" {
		actions = filterActionsByDate(actions, dateFrom, dateTo)
	}
	return map[string]interface{}{
		"status":  "ok",
		"message": "Bug actions retrieved successfully",
		"data": map[string]interface{}{
			"bugId":  bugID,
			"total":  len(actions),
			"counts": countActionsByKind(actions),
			"list":   actions,
		},
	}, nil
}

// filterActionsByDate 按动作日期（YYYY-MM-DD 前缀）过滤；双端全空=不约束原样返回，
// 有约束时空日期行剔除（统计口径诚实——窗口内无日期不计数）。
func filterActionsByDate(actions []myzentao.BugAction, dateFrom, dateTo string) []myzentao.BugAction {
	if dateFrom == "" && dateTo == "" {
		return actions
	}
	filtered := make([]myzentao.BugAction, 0, len(actions))
	for _, a := range actions {
		date := a.Date
		if len(date) > 10 {
			date = date[:10]
		}
		if date == "" {
			continue
		}
		if dateFrom != "" && date < dateFrom {
			continue
		}
		if dateTo != "" && date > dateTo {
			continue
		}
		filtered = append(filtered, a)
	}
	return filtered
}

// countActionsByKind 按 action 类型分组计数（流转统计免二次加工）。
func countActionsByKind(actions []myzentao.BugAction) map[string]int {
	counts := map[string]int{}
	for _, a := range actions {
		kind := a.Action
		if kind == "" {
			kind = "unknown"
		}
		counts[kind]++
	}
	return counts
}
