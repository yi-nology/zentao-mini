package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/models"
)

// defaultCheckHours 每工作日最低工时阈值默认值
const defaultCheckHours = 8.0

// maxMissingDaysInMessage 消息里每人最多列出的未达标日期数，超出部分折叠为"等 N 天"
const maxMissingDaysInMessage = 10

// resolveCheckPeriod 解析检查周期：period 形如 "2026-10"，表示以该月 15 日为截止的周期，
// 即上月 16 日 ~ 该月 15 日（预览时手动指定）；为空时检查最近的周期
// （上月 16 日 ~ 本月 15 日，每月 18 日定时执行的场景）。返回起止日期（含两端）。
func resolveCheckPeriod(period string, now time.Time) (time.Time, time.Time, error) {
	var endYear, endMonth int
	if period == "" {
		endYear, endMonth = now.Year(), int(now.Month())
	} else {
		parsed, err := time.ParseInLocation("2006-01", period, now.Location())
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("检查周期格式应为 YYYY-MM，收到 %q", period)
		}
		endYear, endMonth = parsed.Year(), int(parsed.Month())
	}
	end := time.Date(endYear, time.Month(endMonth), 15, 0, 0, 0, 0, now.Location())
	// time.Date 自动归一化月份下溢：month-1=0 时表示上一年 12 月
	start := time.Date(endYear, time.Month(endMonth)-1, 16, 0, 0, 0, 0, now.Location())
	return start, end, nil
}

// listWorkdays 列出 [start, end] 内的工作日（周一至周五，不含法定节假日调休）；
// limitDay 非零时丢弃不早于其当天零点的日期（当月手动补跑时，当天尚未结束不计应填）。
func listWorkdays(start, end, limitDay time.Time) []time.Time {
	if !limitDay.IsZero() {
		limitDay = time.Date(limitDay.Year(), limitDay.Month(), limitDay.Day(), 0, 0, 0, 0, limitDay.Location())
	}
	days := make([]time.Time, 0, 31)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
			continue
		}
		if !limitDay.IsZero() && !d.Before(limitDay) {
			break
		}
		days = append(days, d)
	}
	return days
}

// GenerateDailyReportCheck 生成产品成员日报完成度检查报告：
// 检查周期内每个工作日每人填报工时是否达到阈值（默认 8h），
// 未达标（含整周期未填报）的人员逐人列出缺失日期。period 为空时检查上月 16 日 ~ 本月 15 日。
func (s *ReportService) GenerateDailyReportCheck(productID int, productName string, checkHours float64, period string, keyword string, externalInfo string, messageHeader string, priorityAssignees []string, viewURL string) (*models.DailyReportCheckReport, error) {
	now := time.Now()
	periodStart, periodEnd, err := resolveCheckPeriod(period, now)
	if err != nil {
		return nil, err
	}
	if checkHours <= 0 {
		checkHours = defaultCheckHours
	}

	if productName == "" {
		if p, err := s.client.GetProduct(productID); err == nil && p != nil {
			productName = p.Name
		}
	}
	if productName == "" {
		productName = fmt.Sprintf("产品%d", productID)
	}

	dateFrom := periodStart.Format("2006-01-02")
	dateTo := periodEnd.Format("2006-01-02")
	summary, err := s.client.GetProductDailyEfforts(productID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("获取产品工时数据失败: %w", err)
	}

	userNames := make(map[string]string)
	if users, err := s.client.GetUsersAll(); err == nil {
		for _, u := range users {
			userNames[u.Account] = u.Realname
		}
	} else {
		logger.Warn("获取用户列表失败，日报检查将直接展示账号", zap.Error(err))
	}

	report := buildDailyCheckReport(now, productID, productName, checkHours, periodStart, periodEnd,
		summary.DailyByAccount, summary.ActiveAssignees, userNames, keyword, externalInfo, messageHeader, priorityAssignees, viewURL)
	return report, nil
}

// buildDailyCheckReport 纯计算：依据按人按日工时汇总判定每个人的日报完成情况并组装消息
func buildDailyCheckReport(
	now time.Time,
	productID int,
	productName string,
	checkHours float64,
	periodStart, periodEnd time.Time,
	dailyByAccount map[string]map[string]float64,
	activeAssignees map[string]bool,
	userNames map[string]string,
	keyword string,
	externalInfo string,
	messageHeader string,
	priorityAssignees []string,
	viewURL string,
) *models.DailyReportCheckReport {
	// 周期为当前月时，今天尚未结束不计应填
	limitDay := now
	if periodEnd.Before(now) {
		limitDay = time.Time{}
	}
	workdays := listWorkdays(periodStart, periodEnd, limitDay)

	displayName := func(account string) string {
		if name := userNames[account]; name != "" {
			return name
		}
		return account
	}

	details := make([]models.AssigneeDailyCheckStats, 0, len(activeAssignees))
	okCount := 0
	totalMissing := 0
	for account := range activeAssignees {
		daily := dailyByAccount[account]
		stat := models.AssigneeDailyCheckStats{
			Assignee:    displayName(account),
			Account:     account,
			Workdays:    len(workdays),
			MissingDays: []models.DailyMissingDay{},
		}
		for _, d := range workdays {
			key := d.Format("2006-01-02")
			hours := daily[key]
			stat.TotalHours += hours
			if hours+1e-9 >= checkHours {
				stat.OkDays++
				continue
			}
			stat.MissingDays = append(stat.MissingDays, models.DailyMissingDay{Date: key, Hours: hours})
		}
		stat.NoEffort = stat.TotalHours <= 0
		if len(stat.MissingDays) > 0 {
			totalMissing += len(stat.MissingDays)
		} else {
			okCount++
		}
		details = append(details, stat)
	}

	prioritySet := make(map[string]bool)
	for _, name := range priorityAssignees {
		prioritySet[name] = true
	}
	sort.Slice(details, func(i, j int) bool {
		iPriority := prioritySet[details[i].Assignee] || prioritySet[details[i].Account]
		jPriority := prioritySet[details[j].Assignee] || prioritySet[details[j].Account]
		if iPriority != jPriority {
			return iPriority
		}
		if len(details[i].MissingDays) != len(details[j].MissingDays) {
			return len(details[i].MissingDays) > len(details[j].MissingDays)
		}
		return details[i].Assignee < details[j].Assignee
	})

	detailURL := buildViewURL(viewURL, "/timelog", map[string]string{
		"product": fmt.Sprintf("%d", productID),
		"from":    periodStart.Format("2006-01-02"),
		"to":      periodEnd.Format("2006-01-02"),
	})

	message := buildDailyCheckMessage(now, productName, checkHours, periodStart, periodEnd, len(workdays),
		details, okCount, totalMissing, keyword, externalInfo, messageHeader, detailURL)

	return &models.DailyReportCheckReport{
		Title:        "日报完成度检查",
		Timestamp:    now.Format(time.RFC3339),
		ProductName:  productName,
		PeriodStart:  periodStart.Format("2006-01-02"),
		PeriodEnd:    periodEnd.Format("2006-01-02"),
		Workdays:     len(workdays),
		CheckHours:   checkHours,
		TotalPeople:  len(details),
		OkCount:      okCount,
		IssueCount:   len(details) - okCount,
		TotalMissing: totalMissing,
		Details:      details,
		Message:      message,
	}
}

// formatMissingDay 未达标日期的消息片段：未填与不足阈值区分展示
func formatMissingDay(m models.DailyMissingDay) string {
	if m.Hours <= 0 {
		return fmt.Sprintf("%s 未填", m.Date[5:])
	}
	return fmt.Sprintf("%s %.1fh", m.Date[5:], m.Hours)
}

func buildDailyCheckMessage(t time.Time, productName string, checkHours float64, periodStart, periodEnd time.Time, workdays int,
	details []models.AssigneeDailyCheckStats, okCount, totalMissing int, keyword string, externalInfo string, messageHeader string, detailURL string) string {
	var sb strings.Builder
	kw := ""
	if keyword != "" {
		kw = fmt.Sprintf("【%s】", keyword)
	}
	sb.WriteString(fmt.Sprintf("%s📝 日报完成度检查 - %s\n", kw, productName))
	sb.WriteString(fmt.Sprintf("📅 %s\n", t.Format("2006-01-02 15:04:05")))
	if messageHeader != "" {
		sb.WriteString(fmt.Sprintf("📌 %s\n", messageHeader))
	}
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("📊 检查周期：%s ~ %s（工作日 %d 天，标准 %.0fh/天，按周一至周五）\n",
		periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"), workdays, checkHours))
	sb.WriteString(fmt.Sprintf("👥 检查 %d 人：达标 %d 人，未达标 %d 人\n", len(details), okCount, len(details)-okCount))
	sb.WriteString(viewLinkLine(detailURL))

	issueIndex := 0
	for _, d := range details {
		if len(d.MissingDays) == 0 {
			continue
		}
		issueIndex++
		if d.NoEffort {
			sb.WriteString(fmt.Sprintf("%d. 👤 %s（%s）周期内无任何工时记录\n", issueIndex, d.Assignee, d.Account))
			continue
		}
		sb.WriteString(fmt.Sprintf("%d. 👤 %s（%s）%d/%d 天达标，缺 %d 天，共 %.1fh\n",
			issueIndex, d.Assignee, d.Account, d.OkDays, d.Workdays, len(d.MissingDays), d.TotalHours))
		shown := d.MissingDays
		if len(shown) > maxMissingDaysInMessage {
			shown = shown[:maxMissingDaysInMessage]
		}
		parts := make([]string, 0, len(shown))
		for _, m := range shown {
			parts = append(parts, formatMissingDay(m))
		}
		sb.WriteString(fmt.Sprintf("   └ %s", strings.Join(parts, " · ")))
		if len(d.MissingDays) > len(shown) {
			sb.WriteString(fmt.Sprintf(" 等共 %d 天", len(d.MissingDays)))
		}
		sb.WriteString("\n")
	}
	if issueIndex == 0 {
		sb.WriteString("✅ 所有人均按工作日完成工时填报，无缺失！\n")
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━\n")
	if externalInfo != "" {
		sb.WriteString(fmt.Sprintf("📌 外部信息：\n%s\n━━━━━━━━━━━━━━━━━━━━\n", externalInfo))
	}
	if totalMissing > 0 {
		sb.WriteString(fmt.Sprintf("⚠️ 共 %d 个人日未达标，请相关人员尽快补填工时！", totalMissing))
	} else {
		sb.WriteString("✅ 本周期日报填报整体合规。")
	}
	return sb.String()
}
