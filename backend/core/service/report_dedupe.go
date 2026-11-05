package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yi-nology/common/biz/zentao"
	"github.com/yi-nology/zentao-mini/backend/core/models"
)

// report_dedupe.go：report_service.go（及 daily 变体）里五组逐字样板的收敛处。
//
// 这些样板的风险不在「占行数」，在「改一份漏四份」：排序语义一变（比如并列
// 时的稳定键换了），五处报告就有两处顺序不一致——用户看到的就是「每次推送
// 排版不一样」。收敛后每条语义只有一份，改一处五处同步。
//
// 兜底：report_service_test.go（消息文案逐字断言）。

// sortAssigneeDetails 排序：优先人员排前 → 数量降序 → 名字稳定（并列时按
// assignee 名排序，保证每次推送顺序一致，方便大家日间对比）。
//
// key 提取排序三元组：常规报告返回 (Assignee, Account, Total)；daily 变体
// 的第二键是缺勤天数，传 len(MissingDays) 即可——排序语义同一条，样本不同。
func sortAssigneeDetails[T any](details []T, priorityAssignees []string, key func(*T) (assignee, account string, total int)) {
	prioritySet := make(map[string]bool)
	for _, name := range priorityAssignees {
		prioritySet[name] = true
	}
	sort.Slice(details, func(i, j int) bool {
		iAssignee, iAccount, iTotal := key(&details[i])
		jAssignee, jAccount, jTotal := key(&details[j])
		iPriority := prioritySet[iAssignee] || prioritySet[iAccount]
		jPriority := prioritySet[jAssignee] || prioritySet[jAccount]
		if iPriority != jPriority {
			return iPriority
		}
		if iTotal != jTotal {
			return iTotal > jTotal
		}
		return iAssignee < jAssignee
	})
}

func bugSortKey(d *models.AssigneeBugStats) (string, string, int) {
	return d.Assignee, d.Account, d.Total
}
func storySortKey(d *models.AssigneeStoryStats) (string, string, int) {
	return d.Assignee, d.Account, d.Total
}
func taskSortKey(d *models.TaskProgressStats) (string, string, int) {
	return d.Assignee, d.Account, d.Total
}
func agingSortKey(d *models.AssigneeBugAgingStats) (string, string, int) {
	return d.Assignee, d.Account, d.Total
}

// writeMessageHeader 报告消息的头框架：关键词角标 + 图标标题 + 时间 +
// 可选置顶说明 + 分隔线。四份报告共用一条排版语义。
func writeMessageHeader(sb *strings.Builder, icon, title, keyword, messageHeader string, t time.Time) {
	kw := ""
	if keyword != "" {
		kw = fmt.Sprintf("【%s】", keyword)
	}
	sb.WriteString(fmt.Sprintf("%s%s %s\n", kw, icon, title))
	sb.WriteString(fmt.Sprintf("📅 %s\n", t.Format("2006-01-02 15:04:05")))
	if messageHeader != "" {
		sb.WriteString(fmt.Sprintf("📌 %s\n", messageHeader))
	}
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
}

// breakdownLabel 状态固定展示顺序与中文标签（避免 map 遍历顺序随机导致
// 每次推送排版不一致）。
type breakdownLabel struct {
	key   string
	label string
}

// formatLabeledBreakdown 按固定顺序输出状态分布；未收录但计数>0 的状态按
// 名字排序追加在末尾（新状态不丢，只是排在已收录档之后）。
func formatLabeledBreakdown(statusBreakdown map[string]int, labels []breakdownLabel, emptyText string) string {
	parts := make([]string, 0, len(statusBreakdown))
	seen := map[string]bool{}
	for _, s := range labels {
		if c := statusBreakdown[s.key]; c > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", s.label, c))
			seen[s.key] = true
		}
	}
	keys := make([]string, 0, len(statusBreakdown))
	for k := range statusBreakdown {
		if !seen[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		if c := statusBreakdown[k]; c > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", k, c))
		}
	}
	if len(parts) == 0 {
		return emptyText
	}
	return strings.Join(parts, " | ")
}

// formatNonzeroCounts 明细行：固定顺序、只显示非零档。
func formatNonzeroCounts(pairs []struct {
	n int
	s string
}, emptyText string) string {
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		if p.n > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", p.s, p.n))
		}
	}
	if len(parts) == 0 {
		return emptyText
	}
	return strings.Join(parts, " ")
}

// assignedToNameAccount 提取 AssignedTo 的展示名与账号：平台把它序列化成
// UserRef 或裸 map（v5 两种应答形状都见过），读取侧统一在这里收口。
// 名字为空时按惯例给「未指派」。
func assignedToNameAccount(v interface{}) (string, string) {
	name := ""
	account := ""
	if ref, ok := v.(zentao.UserRef); ok {
		name = ref.Realname
		account = ref.Account
	} else if ref, ok := v.(map[string]interface{}); ok {
		if s, ok := ref["realname"].(string); ok {
			name = s
		}
		if s, ok := ref["account"].(string); ok {
			account = s
		}
	}
	if name == "" {
		name = "未指派"
	}
	return name, account
}
