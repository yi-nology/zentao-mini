package service

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yi-nology/common/biz/zentao"
	"go.uber.org/zap"

	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/models"
	myzentao "github.com/yi-nology/zentao-mini/backend/core/zentao"
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

// buildViewURL 基于回访地址（zentao-mini 站点，如 https://zentao.kylin.me）拼接带过滤参数的页面深链；
// viewURL 为空时返回空串，消息保持原有格式。
func buildViewURL(viewURL string, path string, params map[string]string) string {
	base := strings.TrimSuffix(strings.TrimSpace(viewURL), "/")
	if base == "" {
		return ""
	}
	full := base + path
	values := url.Values{}
	for k, v := range params {
		if v != "" {
			values.Set(k, v)
		}
	}
	if encoded := values.Encode(); encoded != "" {
		full += "?" + encoded
	}
	return full
}

// detailLinkParams 组装页面查询参数：product/project 二选一或同时，status 过滤 "all"/"active-resolved" 无法在页面表达，忽略。
func detailLinkParams(productID, projectID int, statusFilter string) map[string]string {
	params := map[string]string{}
	if productID > 0 {
		params["product"] = strconv.Itoa(productID)
	}
	if projectID > 0 {
		params["project"] = strconv.Itoa(projectID)
	}
	if statusFilter != "" && statusFilter != "all" && statusFilter != "active-resolved" {
		params["status"] = statusFilter
	}
	return params
}

// viewLinkLine 消息中的"查看详情"链接行；link 为空时返回空串。
func viewLinkLine(link string) string {
	if link == "" {
		return ""
	}
	return fmt.Sprintf("🔗 查看详情：%s\n", link)
}

// zentaoBugURL 禅道 Bug 详情页地址（与前端 bug-view-{id}.html 跳转规则一致）；服务未配置时返回空串。
func zentaoBugURL(zentaoBase string, bugID int) string {
	base := strings.TrimSuffix(strings.TrimSpace(zentaoBase), "/")
	if base == "" {
		return ""
	}
	return fmt.Sprintf("%s/bug-view-%d.html", base, bugID)
}

func (s *ReportService) GenerateBugReport(productID int, projectID int, projectName string, statusFilter string, keyword string, externalInfo string, messageHeader string, priorityAssignees []string, viewURL string) (*models.BugReport, error) {
	bugs, err := s.client.GetAllBugsByProjectWithProduct(productID, projectID)
	if err != nil {
		return nil, fmt.Errorf("获取Bug列表失败: %w", err)
	}

	logger.Info("获取Bug列表",
		zap.Int("productID", productID),
		zap.Int("projectID", projectID),
		zap.Int("bugCount", len(bugs)))

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

	logger.Info("Bug过滤结果",
		zap.String("statusFilter", statusFilter),
		zap.Int("filteredCount", len(filtered)),
		zap.Any("statusBreakdown", statusBreakdown))

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
		case 4:
			stat.Minor++
		case 5:
			stat.Suggest++
		}
	}

	details := make([]models.AssigneeBugStats, 0, len(assigneeMap))
	for _, stat := range assigneeMap {
		details = append(details, *stat)
	}

	// 构建优先人员集合
	prioritySet := make(map[string]bool)
	for _, name := range priorityAssignees {
		prioritySet[name] = true
	}

	// 排序：优先人员排前面，然后按 Bug 数量降序
	sort.Slice(details, func(i, j int) bool {
		iPriority := prioritySet[details[i].Assignee] || prioritySet[details[i].Account]
		jPriority := prioritySet[details[j].Assignee] || prioritySet[details[j].Account]
		if iPriority != jPriority {
			return iPriority
		}
		if details[i].Total != details[j].Total {
			return details[i].Total > details[j].Total
		}
		// 数量并列时按名字排序，保证每次推送顺序稳定，方便大家日间对比
		return details[i].Assignee < details[j].Assignee
	})

	totalHigh := 0
	for _, d := range details {
		totalHigh += d.HighSeverity
	}

	now := time.Now()
	title := fmt.Sprintf("Bug 分布报告 - %s", projectName)
	detailURL := buildViewURL(viewURL, "/bugs", detailLinkParams(productID, projectID, statusFilter))
	message := buildMessage(title, now, len(filtered), statusFilter, details, statusBreakdown, keyword, externalInfo, messageHeader, detailURL)

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

// bugStatusLabel 状态过滤条件的中文标签，用于汇总行说清统计口径
func bugStatusLabel(statusFilter string) string {
	switch statusFilter {
	case "active", "":
		return "活跃 Bug"
	case "resolved":
		return "已解决 Bug"
	case "closed":
		return "已关闭 Bug"
	case "active-resolved":
		return "活跃+已解决 Bug"
	case "all":
		return "Bug"
	default:
		return fmt.Sprintf("%s Bug", statusFilter)
	}
}

// formatSeverityBreakdown 人员严重度明细：固定顺序、只显示非零档
func formatSeverityBreakdown(d models.AssigneeBugStats) string {
	parts := make([]string, 0, 5)
	for _, p := range []struct {
		n int
		s string
	}{
		{d.Fatal, "致命"}, {d.Serious, "严重"}, {d.Moderate, "一般"}, {d.Minor, "轻微"}, {d.Suggest, "建议"},
	} {
		if p.n > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", p.s, p.n))
		}
	}
	if len(parts) == 0 {
		return "未分级"
	}
	return strings.Join(parts, " ")
}

func buildMessage(title string, t time.Time, total int, statusFilter string, details []models.AssigneeBugStats, statusBreakdown map[string]int, keyword string, externalInfo string, messageHeader string, detailURL string) string {
	var sb strings.Builder
	kw := ""
	if keyword != "" {
		kw = fmt.Sprintf("【%s】", keyword)
	}
	sb.WriteString(fmt.Sprintf("%s🔴 %s\n", kw, title))
	sb.WriteString(fmt.Sprintf("📅 %s\n", t.Format("2006-01-02 15:04:05")))
	if messageHeader != "" {
		sb.WriteString(fmt.Sprintf("📌 %s\n", messageHeader))
	}
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")

	totalFatal, totalSerious, totalSuggest := 0, 0, 0
	for _, d := range details {
		totalFatal += d.Fatal
		totalSerious += d.Serious
		totalSuggest += d.Suggest
	}
	totalCritical := totalFatal + totalSerious

	sb.WriteString(fmt.Sprintf("📊 %s：%d个，严重级别（致命+严重）%d个\n", bugStatusLabel(statusFilter), total, totalCritical))
	sb.WriteString(viewLinkLine(detailURL))
	sb.WriteString("\n")

	for _, d := range details {
		seriousStr := ""
		if serious := d.Fatal + d.Serious; serious > 0 {
			seriousStr = fmt.Sprintf("，严重 %d个", serious)
		}
		sb.WriteString(fmt.Sprintf("👤 %s  %d个%s\n", d.Assignee, d.Total, seriousStr))
		sb.WriteString(fmt.Sprintf("   └ %s\n", formatSeverityBreakdown(d)))
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━\n")
	if externalInfo != "" {
		sb.WriteString(fmt.Sprintf("📌 外部信息：\n%s\n━━━━━━━━━━━━━━━━━━━━\n", externalInfo))
	}
	if totalCritical > 0 {
		sb.WriteString(fmt.Sprintf("⚠️ 严重级别共 %d个（致命 %d | 严重 %d），需重点关注！\n", totalCritical, totalFatal, totalSerious))
	}
	if totalSuggest > 0 {
		sb.WriteString(fmt.Sprintf("💡 另有建议级 %d个\n", totalSuggest))
	}
	sb.WriteString(fmt.Sprintf("📈 状态分布：活跃 %d | 已解决 %d | 已关闭 %d",
		statusBreakdown["active"], statusBreakdown["resolved"], statusBreakdown["closed"]))
	return sb.String()
}

func (s *ReportService) GenerateRequirementReport(productID int, projectID int, projectName string, productName string, keyword string, externalInfo string, messageHeader string, priorityAssignees []string, viewURL string) (*models.RequirementReport, error) {
	var stories []zentao.Story
	var err error

	if projectID > 0 {
		stories, err = s.client.GetAllStoriesByProject(projectID)
	} else if productID > 0 {
		stories, err = s.client.GetAllStories(productID)
	} else {
		return nil, fmt.Errorf("请提供产品ID或项目ID")
	}
	if err != nil {
		return nil, fmt.Errorf("获取需求列表失败: %w", err)
	}

	statusBreakdown := map[string]int{}
	for _, st := range stories {
		statusBreakdown[st.Status]++
	}

	assigneeMap := make(map[string]*models.AssigneeStoryStats)
	for _, st := range stories {
		name := ""
		account := ""
		if ref, ok := st.AssignedTo.(zentao.UserRef); ok {
			name = ref.Realname
			account = ref.Account
		} else if ref, ok := st.AssignedTo.(map[string]interface{}); ok {
			if v, ok := ref["realname"].(string); ok {
				name = v
			}
			if v, ok := ref["account"].(string); ok {
				account = v
			}
		}
		if name == "" {
			name = "未指派"
		}
		stat, ok := assigneeMap[name]
		if !ok {
			stat = &models.AssigneeStoryStats{
				Assignee: name,
				Account:  account,
			}
			assigneeMap[name] = stat
		}
		stat.Total++
		switch st.Status {
		case "active":
			stat.Active++
		case "changed":
			stat.Changed++
		case "closed":
			stat.Closed++
		case "resolved":
			stat.Resolved++
		case "accepted":
			stat.Accepted++
		case "reviewing":
			stat.Reviewing++
		}
	}

	details := make([]models.AssigneeStoryStats, 0, len(assigneeMap))
	for _, stat := range assigneeMap {
		details = append(details, *stat)
	}

	// 构建优先人员集合
	prioritySet := make(map[string]bool)
	for _, name := range priorityAssignees {
		prioritySet[name] = true
	}

	// 排序：优先人员排前面，然后按需求数量降序
	sort.Slice(details, func(i, j int) bool {
		iPriority := prioritySet[details[i].Assignee] || prioritySet[details[i].Account]
		jPriority := prioritySet[details[j].Assignee] || prioritySet[details[j].Account]
		if iPriority != jPriority {
			return iPriority
		}
		if details[i].Total != details[j].Total {
			return details[i].Total > details[j].Total
		}
		// 数量并列时按名字排序，保证每次推送顺序稳定，方便大家日间对比
		return details[i].Assignee < details[j].Assignee
	})

	now := time.Now()
	title := fmt.Sprintf("需求进度报告 - %s", projectName)
	detailURL := buildViewURL(viewURL, "/stories", detailLinkParams(productID, projectID, ""))
	message := buildRequirementMessage(title, now, len(stories), details, statusBreakdown, keyword, externalInfo, messageHeader, detailURL)

	return &models.RequirementReport{
		Title:           title,
		Timestamp:       now.Format(time.RFC3339),
		ProjectName:     projectName,
		ProductName:     productName,
		Total:           len(stories),
		StatusBreakdown: statusBreakdown,
		Details:         details,
		Message:         message,
	}, nil
}

// storyStatusLabels 需求状态固定展示顺序与中文标签，避免 map 遍历顺序随机导致每次推送排版不一致。
var storyStatusLabels = []struct {
	key   string
	label string
}{
	{"active", "激活"},
	{"reviewing", "评审中"},
	{"changed", "已变更"},
	{"resolved", "已解决"},
	{"accepted", "已验收"},
	{"closed", "已关闭"},
	{"draft", "草稿"},
}

// formatStoryBreakdown 按固定顺序输出需求状态分布，未收录但计数>0 的状态追加在末尾。
func formatStoryBreakdown(statusBreakdown map[string]int) string {
	parts := make([]string, 0, len(statusBreakdown))
	seen := map[string]bool{}
	for _, s := range storyStatusLabels {
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
		return "暂无需求"
	}
	return strings.Join(parts, " | ")
}

// formatStoryDetail 人员需求状态明细：只显示非零档
func formatStoryDetail(d models.AssigneeStoryStats) string {
	parts := make([]string, 0, 6)
	for _, p := range []struct {
		n int
		s string
	}{
		{d.Active, "激活"}, {d.Reviewing, "评审中"}, {d.Changed, "变更"}, {d.Resolved, "已解决"}, {d.Accepted, "已验收"}, {d.Closed, "已关闭"},
	} {
		if p.n > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", p.s, p.n))
		}
	}
	if len(parts) == 0 {
		return "无状态明细"
	}
	return strings.Join(parts, " ")
}

func buildRequirementMessage(title string, t time.Time, total int, details []models.AssigneeStoryStats, statusBreakdown map[string]int, keyword string, externalInfo string, messageHeader string, detailURL string) string {
	var sb strings.Builder
	kw := ""
	if keyword != "" {
		kw = fmt.Sprintf("【%s】", keyword)
	}
	sb.WriteString(fmt.Sprintf("%s📋 %s\n", kw, title))
	sb.WriteString(fmt.Sprintf("📅 %s\n", t.Format("2006-01-02 15:04:05")))
	if messageHeader != "" {
		sb.WriteString(fmt.Sprintf("📌 %s\n", messageHeader))
	}
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("📊 需求总数：%d个\n", total))
	sb.WriteString(viewLinkLine(detailURL))
	sb.WriteString("\n")

	for _, d := range details {
		sb.WriteString(fmt.Sprintf("👤 %s  共%d个需求\n", d.Assignee, d.Total))
		sb.WriteString(fmt.Sprintf("   └ %s\n", formatStoryDetail(d)))
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━\n")
	if externalInfo != "" {
		sb.WriteString(fmt.Sprintf("📌 外部信息：\n%s\n━━━━━━━━━━━━━━━━━━━━\n", externalInfo))
	}
	sb.WriteString(fmt.Sprintf("📈 需求状态分布：%s", formatStoryBreakdown(statusBreakdown)))
	return sb.String()
}

func (s *ReportService) GenerateTaskReport(productID int, projectID int, projectName string, productName string, keyword string, externalInfo string, messageHeader string, priorityAssignees []string, viewURL string) (*models.TaskProgressReport, error) {
	var tasks []zentao.Task
	var err error

	if projectID > 0 {
		tasks, err = s.client.GetAllTasksByProject(projectID)
	} else if productID > 0 {
		tasks, err = s.client.GetAllTasksByProduct(productID)
	} else {
		return nil, fmt.Errorf("请提供产品ID或项目ID")
	}
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %w", err)
	}

	statusBreakdown := map[string]int{}
	var totalEstimate, totalConsumed float64

	for _, t := range tasks {
		statusBreakdown[t.Status]++
		totalEstimate += toFloat64(t.Estimate)
		totalConsumed += toFloat64(t.Consumed)
	}

	var overallProgress float64
	if totalEstimate > 0 {
		overallProgress = (totalConsumed / totalEstimate) * 100
		if overallProgress > 100 {
			overallProgress = 100
		}
	}

	assigneeMap := make(map[string]*models.TaskProgressStats)
	for _, t := range tasks {
		name := ""
		account := ""
		if ref, ok := t.AssignedTo.(zentao.UserRef); ok {
			name = ref.Realname
			account = ref.Account
		} else if ref, ok := t.AssignedTo.(map[string]interface{}); ok {
			if v, ok := ref["realname"].(string); ok {
				name = v
			}
			if v, ok := ref["account"].(string); ok {
				account = v
			}
		}
		if name == "" {
			name = "未指派"
		}
		stat, ok := assigneeMap[name]
		if !ok {
			stat = &models.TaskProgressStats{
				Assignee: name,
				Account:  account,
			}
			assigneeMap[name] = stat
		}
		stat.Total++
		stat.Estimate += toFloat64(t.Estimate)
		stat.Consumed += toFloat64(t.Consumed)
		switch t.Status {
		case "wait":
			stat.Wait++
		case "doing":
			stat.Doing++
		case "done":
			stat.Done++
		case "pause":
			stat.Paused++
		case "cancel":
			stat.Cancelled++
		}
	}

	for _, stat := range assigneeMap {
		if stat.Estimate > 0 {
			stat.Progress = (stat.Consumed / stat.Estimate) * 100
			if stat.Progress > 100 {
				stat.Progress = 100
			}
		}
	}

	details := make([]models.TaskProgressStats, 0, len(assigneeMap))
	for _, stat := range assigneeMap {
		details = append(details, *stat)
	}

	// 构建优先人员集合
	prioritySet := make(map[string]bool)
	for _, name := range priorityAssignees {
		prioritySet[name] = true
	}

	// 排序：优先人员排前面，然后按任务数量降序
	sort.Slice(details, func(i, j int) bool {
		iPriority := prioritySet[details[i].Assignee] || prioritySet[details[i].Account]
		jPriority := prioritySet[details[j].Assignee] || prioritySet[details[j].Account]
		if iPriority != jPriority {
			return iPriority
		}
		if details[i].Total != details[j].Total {
			return details[i].Total > details[j].Total
		}
		// 数量并列时按名字排序，保证每次推送顺序稳定，方便大家日间对比
		return details[i].Assignee < details[j].Assignee
	})

	now := time.Now()
	title := fmt.Sprintf("任务进度报告 - %s", projectName)
	detailURL := buildViewURL(viewURL, "/tasks", detailLinkParams(productID, projectID, ""))
	message := buildTaskMessage(title, now, len(tasks), totalEstimate, totalConsumed, overallProgress, details, statusBreakdown, keyword, externalInfo, messageHeader, detailURL)

	return &models.TaskProgressReport{
		Title:           title,
		Timestamp:       now.Format(time.RFC3339),
		ProjectName:     projectName,
		ProductName:     productName,
		Total:           len(tasks),
		StatusBreakdown: statusBreakdown,
		TotalEstimate:   totalEstimate,
		TotalConsumed:   totalConsumed,
		OverallProgress: overallProgress,
		Details:         details,
		Message:         message,
	}, nil
}

// taskStatusLabels 任务状态固定展示顺序与中文标签（含旧数据可能出现的 closed）。
var taskStatusLabels = []struct {
	key   string
	label string
}{
	{"wait", "待开始"},
	{"doing", "进行中"},
	{"done", "已完成"},
	{"pause", "已暂停"},
	{"cancel", "已取消"},
	{"closed", "已关闭"},
}

// formatTaskBreakdown 按固定顺序输出任务状态分布。
func formatTaskBreakdown(statusBreakdown map[string]int) string {
	parts := make([]string, 0, len(statusBreakdown))
	seen := map[string]bool{}
	for _, s := range taskStatusLabels {
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
		return "暂无任务"
	}
	return strings.Join(parts, " | ")
}

// formatTaskDetail 人员任务状态明细：只显示非零档
func formatTaskDetail(d models.TaskProgressStats) string {
	parts := make([]string, 0, 5)
	for _, p := range []struct {
		n int
		s string
	}{
		{d.Wait, "待开始"}, {d.Doing, "进行中"}, {d.Done, "已完成"}, {d.Paused, "已暂停"}, {d.Cancelled, "已取消"},
	} {
		if p.n > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", p.s, p.n))
		}
	}
	if len(parts) == 0 {
		return "无状态明细"
	}
	return strings.Join(parts, " ")
}

func buildTaskMessage(title string, t time.Time, total int, totalEstimate, totalConsumed, overallProgress float64, details []models.TaskProgressStats, statusBreakdown map[string]int, keyword string, externalInfo string, messageHeader string, detailURL string) string {
	var sb strings.Builder
	kw := ""
	if keyword != "" {
		kw = fmt.Sprintf("【%s】", keyword)
	}
	sb.WriteString(fmt.Sprintf("%s✅ %s\n", kw, title))
	sb.WriteString(fmt.Sprintf("📅 %s\n", t.Format("2006-01-02 15:04:05")))
	if messageHeader != "" {
		sb.WriteString(fmt.Sprintf("📌 %s\n", messageHeader))
	}
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("📊 任务总数：%d个 | 整体进度：%.0f%%\n", total, overallProgress))
	sb.WriteString(fmt.Sprintf("⏱ 预估工时：%.1fh | 已消耗：%.1fh\n", totalEstimate, totalConsumed))
	sb.WriteString(viewLinkLine(detailURL))
	sb.WriteString("\n")

	for _, d := range details {
		progressStr := fmt.Sprintf("%.0f%%", d.Progress)
		sb.WriteString(fmt.Sprintf("👤 %s  共%d个任务  进度%s\n", d.Assignee, d.Total, progressStr))
		sb.WriteString(fmt.Sprintf("   └ %s\n", formatTaskDetail(d)))
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━\n")
	if externalInfo != "" {
		sb.WriteString(fmt.Sprintf("📌 外部信息：\n%s\n━━━━━━━━━━━━━━━━━━━━\n", externalInfo))
	}
	sb.WriteString(fmt.Sprintf("📈 任务状态分布：%s", formatTaskBreakdown(statusBreakdown)))
	return sb.String()
}

func (s *ReportService) GenerateBugAgingReport(productID int, projectID int, projectName string, statusFilter string, agingDays int, keyword string, externalInfo string, priorityAssignees []string, messageHeader string, viewURL string) (*models.BugAgingReport, error) {
	bugs, err := s.client.GetAllBugsByProjectWithProduct(productID, projectID)
	if err != nil {
		return nil, fmt.Errorf("获取Bug列表失败: %w", err)
	}

	if agingDays <= 0 {
		agingDays = 7
	}

	logger.Info("获取Bug列表(超时分析)",
		zap.Int("productID", productID),
		zap.Int("projectID", projectID),
		zap.Int("bugCount", len(bugs)),
		zap.Int("agingDays", agingDays))

	if statusFilter == "" {
		statusFilter = "active"
	}

	now := time.Now()
	assigneeMap := make(map[string]*models.AssigneeBugAgingStats)

	for _, b := range bugs {
		if b.Status != "active" && b.Status != "resolved" {
			continue
		}
		if statusFilter != "all" && b.Status != statusFilter && statusFilter != "active-resolved" {
			continue
		}

		openedStr := parseDateField(b.OpenedDate)
		if openedStr == "" {
			continue
		}
		openedTime, err := time.ParseInLocation("2006-01-02T15:04:05Z", openedStr, time.UTC)
		if err != nil {
			openedTime, err = time.ParseInLocation("2006-01-02 15:04:05", openedStr, time.Local)
			if err != nil {
				openedTime, err = time.ParseInLocation("2006-01-02", openedStr, time.Local)
				if err != nil {
					continue
				}
			}
		}
		daysOpen := int(now.Sub(openedTime).Hours() / 24)
		if daysOpen < agingDays {
			continue
		}

		name := b.AssignedTo.Realname
		if name == "" {
			name = b.AssignedTo.Account
		}
		if name == "" {
			name = "未指派"
		}

		stat, ok := assigneeMap[name]
		if !ok {
			stat = &models.AssigneeBugAgingStats{
				Assignee: name,
				Account:  b.AssignedTo.Account,
				Bugs:     []models.BugAgingItem{},
			}
			assigneeMap[name] = stat
		}

		sev := severityInt(b.Severity)
		sevName := severityName(sev)

		stat.Total++
		stat.Bugs = append(stat.Bugs, models.BugAgingItem{
			ID:         b.ID,
			Title:      b.Title,
			Severity:   sevName,
			OpenedDate: openedStr,
			DaysOpen:   daysOpen,
		})
	}

	details := make([]models.AssigneeBugAgingStats, 0, len(assigneeMap))
	for _, stat := range assigneeMap {
		sort.Slice(stat.Bugs, func(i, j int) bool {
			return stat.Bugs[i].DaysOpen > stat.Bugs[j].DaysOpen
		})
		details = append(details, *stat)
	}

	// 构建优先人员集合
	prioritySet := make(map[string]bool)
	for _, name := range priorityAssignees {
		prioritySet[name] = true
	}

	// 排序：优先人员排前面，然后按 Bug 数量降序
	sort.Slice(details, func(i, j int) bool {
		iPriority := prioritySet[details[i].Assignee] || prioritySet[details[i].Account]
		jPriority := prioritySet[details[j].Assignee] || prioritySet[details[j].Account]
		if iPriority != jPriority {
			return iPriority
		}
		if details[i].Total != details[j].Total {
			return details[i].Total > details[j].Total
		}
		// 数量并列时按名字排序，保证每次推送顺序稳定，方便大家日间对比
		return details[i].Assignee < details[j].Assignee
	})

	total := 0
	for _, d := range details {
		total += d.Total
	}

	title := fmt.Sprintf("Bug 停留超时提醒 - %s", projectName)
	detailURL := buildViewURL(viewURL, "/bugs", detailLinkParams(productID, projectID, statusFilter))
	message := buildBugAgingMessage(title, now, total, agingDays, details, keyword, externalInfo, messageHeader, detailURL, s.client.GetServer())

	return &models.BugAgingReport{
		Title:       title,
		Timestamp:   now.Format(time.RFC3339),
		ProjectName: projectName,
		Total:       total,
		AgingDays:   agingDays,
		Details:     details,
		Message:     message,
	}, nil
}

func parseDateField(v interface{}) string {
	switch d := v.(type) {
	case string:
		return d
	case time.Time:
		return d.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func severityName(sev int) string {
	switch sev {
	case 1:
		return "致命"
	case 2:
		return "严重"
	case 3:
		return "一般"
	case 4:
		return "轻微"
	case 5:
		return "建议"
	default:
		return "未知"
	}
}

// agingMaxListPerPerson 每人最多列出的超时 Bug 条数，超出折叠提示（避免长消息刷屏）
const agingMaxListPerPerson = 5

func buildBugAgingMessage(title string, t time.Time, total, agingDays int, details []models.AssigneeBugAgingStats, keyword string, externalInfo string, messageHeader string, detailURL string, zentaoBase string) string {
	var sb strings.Builder
	kw := ""
	if keyword != "" {
		kw = fmt.Sprintf("【%s】", keyword)
	}
	sb.WriteString(fmt.Sprintf("%s⏰ %s\n", kw, title))
	sb.WriteString(fmt.Sprintf("📅 %s\n", t.Format("2006-01-02 15:04:05")))
	if messageHeader != "" {
		sb.WriteString(fmt.Sprintf("📌 %s\n", messageHeader))
	}
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")

	maxDays := 0
	for _, d := range details {
		for _, bug := range d.Bugs {
			if bug.DaysOpen > maxDays {
				maxDays = bug.DaysOpen
			}
		}
	}
	stayStr := ""
	if maxDays > 0 {
		stayStr = fmt.Sprintf("，最长已停留 %d天", maxDays)
	}
	sb.WriteString(fmt.Sprintf("📊 超时 Bug：%d个（阈值：%d天%s）\n", total, agingDays, stayStr))
	sb.WriteString(viewLinkLine(detailURL))
	sb.WriteString("\n")

	for _, d := range details {
		sb.WriteString(fmt.Sprintf("👤 %s  %d个超时Bug\n", d.Assignee, d.Total))
		shown := 0
		for _, bug := range d.Bugs {
			if shown >= agingMaxListPerPerson {
				rest := d.Total - shown
				sb.WriteString(fmt.Sprintf("   └ …等 %d 条未列出，详见上方链接\n", rest))
				break
			}
			titleStr := bug.Title
			if len([]rune(titleStr)) > 20 {
				titleStr = string([]rune(titleStr)[:20]) + "..."
			}
			link := zentaoBugURL(zentaoBase, bug.ID)
			if link != "" {
				sb.WriteString(fmt.Sprintf("   └ [%s] %s 已停留 %d天\n      🔗 %s\n", bug.Severity, titleStr, bug.DaysOpen, link))
			} else {
				sb.WriteString(fmt.Sprintf("   └ #%d [%s] %s 已停留 %d天\n", bug.ID, bug.Severity, titleStr, bug.DaysOpen))
			}
			shown++
		}
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━\n")
	if externalInfo != "" {
		sb.WriteString(fmt.Sprintf("📌 外部信息：\n%s\n━━━━━━━━━━━━━━━━━━━━\n", externalInfo))
	}
	sb.WriteString(fmt.Sprintf("⚠️ 以上 Bug 已超过 %d 天未解决，请尽快处理！", agingDays))
	return sb.String()
}
