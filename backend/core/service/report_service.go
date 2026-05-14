package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yi-nology/common/biz/zentao"

	"chandao-mini/backend/core/models"
	myzentao "chandao-mini/backend/core/zentao"
)

type ReportService struct {
	client *myzentao.Client
}

func NewReportService(client *myzentao.Client) *ReportService {
	return &ReportService{client: client}
}

func severityInt(v interface{}) int {
	switch s := v.(type) {
	case float64:
		return int(s)
	case int:
		return s
	case string:
		var n int
		fmt.Sscanf(s, "%d", &n)
		return n
	default:
		return 0
	}
}

func (s *ReportService) GenerateBugReport(projectID int, projectName string, statusFilter string) (*models.BugReport, error) {
	bugs, err := s.client.GetAllBugsByProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("获取Bug列表失败: %w", err)
	}

	if statusFilter == "" {
		statusFilter = "active"
	}

	filtered := make([]zentao.Bug, 0)
	statusBreakdown := map[string]int{"active": 0, "resolved": 0, "closed": 0}
	for _, b := range bugs {
		statusBreakdown[b.Status]++
		if statusFilter == "all" || b.Status == statusFilter {
			filtered = append(filtered, b)
		}
	}

	assigneeMap := make(map[string]*models.AssigneeBugStats)
	for _, b := range filtered {
		name := b.AssignedTo.Realname
		if name == "" {
			name = b.AssignedTo.Account
		}
		if name == "" {
			name = "未指派"
		}
		stat, ok := assigneeMap[name]
		if !ok {
			stat = &models.AssigneeBugStats{
				Assignee: name,
				Account:  b.AssignedTo.Account,
			}
			assigneeMap[name] = stat
		}
		stat.Total++
		sev := severityInt(b.Severity)
		switch sev {
		case 1:
			stat.Fatal++
			stat.HighSeverity++
		case 2:
			stat.Serious++
			stat.HighSeverity++
		case 3:
			stat.Moderate++
			stat.HighSeverity++
		case 4:
			stat.Minor++
		}
	}

	details := make([]models.AssigneeBugStats, 0, len(assigneeMap))
	for _, stat := range assigneeMap {
		details = append(details, *stat)
	}
	sort.Slice(details, func(i, j int) bool {
		return details[i].Total > details[j].Total
	})

	totalHigh := 0
	for _, d := range details {
		totalHigh += d.HighSeverity
	}

	now := time.Now()
	title := fmt.Sprintf("Bug 分布报告 - %s", projectName)
	message := buildMessage(title, now, len(filtered), totalHigh, details, statusBreakdown)

	return &models.BugReport{
		Title:           title,
		Timestamp:       now.Format(time.RFC3339),
		ProjectName:     projectName,
		Total:           len(filtered),
		HighSeverity:    totalHigh,
		StatusBreakdown: statusBreakdown,
		Details:         details,
		Message:         message,
	}, nil
}

func buildMessage(title string, t time.Time, total, highSeverity int, details []models.AssigneeBugStats, statusBreakdown map[string]int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("【提醒】🔴 %s\n", title))
	sb.WriteString(fmt.Sprintf("📅 %s\n", t.Format("2006-01-02 15:04:05")))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("📊 剩余 Bug：%d个（高级别 %d个）\n\n", total, highSeverity))

	for _, d := range details {
		highStr := ""
		if d.HighSeverity > 0 {
			highStr = fmt.Sprintf(" 其中%d个高级别", d.HighSeverity)
		}
		sb.WriteString(fmt.Sprintf("👤 %s  %d个%s\n", d.Assignee, d.Total, highStr))
		sb.WriteString(fmt.Sprintf("   └ 致命:%d 严重:%d 一般:%d 轻微:%d\n", d.Fatal, d.Serious, d.Moderate, d.Minor))
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("⚠️ 高级别 Bug 共 %d个，需重点关注！\n", highSeverity))
	sb.WriteString(fmt.Sprintf("📈 Bug 状态分布：活跃 %d | 已解决 %d | 已关闭 %d",
		statusBreakdown["active"], statusBreakdown["resolved"], statusBreakdown["closed"]))
	return sb.String()
}
