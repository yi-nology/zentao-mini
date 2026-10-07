package service

import (
	"strings"
	"testing"
	"time"

	"github.com/yi-nology/zentao-mini/backend/core/models"
)

func TestResolveCheckPeriod(t *testing.T) {
	now := time.Date(2026, 10, 18, 9, 0, 0, 0, time.Local)

	start, end, err := resolveCheckPeriod("", now)
	if err != nil {
		t.Fatalf("空周期解析失败: %v", err)
	}
	if start.Format("2006-01-02") != "2026-09-16" || end.Format("2006-01-02") != "2026-10-15" {
		t.Fatalf("空周期应检查上月16日~本月15日 2026-09-16~2026-10-15，得到 %s ~ %s", start, end)
	}

	// 指定月份表示周期截止月：2026-02 → 2026-01-16 ~ 2026-02-15
	start, end, err = resolveCheckPeriod("2026-02", now)
	if err != nil {
		t.Fatalf("指定周期解析失败: %v", err)
	}
	if start.Format("2006-01-02") != "2026-01-16" || end.Format("2006-01-02") != "2026-02-15" {
		t.Fatalf("2026-02 应为 01-16 ~ 02-15，得到 %s ~ %s", start, end)
	}

	// 跨年：2026-01 → 2025-12-16 ~ 2026-01-15
	start, end, err = resolveCheckPeriod("2026-01", now)
	if err != nil {
		t.Fatalf("跨年周期解析失败: %v", err)
	}
	if start.Format("2006-01-02") != "2025-12-16" || end.Format("2006-01-02") != "2026-01-15" {
		t.Fatalf("2026-01 应为 2025-12-16 ~ 2026-01-15，得到 %s ~ %s", start, end)
	}

	if _, _, err := resolveCheckPeriod("2026-9", now); err == nil {
		t.Fatal("非法周期格式应返回错误")
	}
}

func TestListWorkdays(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local) // 周二
	end := time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local)

	days := listWorkdays(start, end, time.Time{})
	if len(days) != 22 {
		t.Fatalf("2026-09 应有 22 个工作日，得到 %d", len(days))
	}
	for _, d := range days {
		if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
			t.Fatalf("工作日列表不应包含周末: %s", d)
		}
	}

	// 当月手动补跑：只统计到 limitDay 之前（当天未结束不计应填）
	limit := time.Date(2026, 9, 10, 12, 0, 0, 0, time.Local)
	days = listWorkdays(start, end, limit)
	if len(days) != 7 { // 9-01(二) ~ 9-09(三)，剔除周末后 7 天
		t.Fatalf("截止 09-10 应有 7 个工作日，得到 %d", len(days))
	}
	if days[len(days)-1].Format("2006-01-02") != "2026-09-09" {
		t.Fatalf("最后一个工作日应为 2026-09-09，得到 %s", days[len(days)-1])
	}
}

func dailyMap(pairs ...interface{}) map[string]float64 {
	m := make(map[string]float64)
	for i := 0; i+1 < len(pairs); i += 2 {
		m[pairs[i].(string)] = pairs[i+1].(float64)
	}
	return m
}

func TestBuildDailyCheckReport(t *testing.T) {
	now := time.Date(2026, 10, 18, 9, 0, 0, 0, time.Local)
	periodStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local)
	periodEnd := time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local)

	// 2026-09 有 22 个工作日；09-07、09-14、09-21、09-28 为周一
	daily := map[string]map[string]float64{
		"zhangsan": {
			"2026-09-01": 8, "2026-09-02": 9, "2026-09-03": 8, "2026-09-04": 8,
			"2026-09-07": 8, "2026-09-08": 8, "2026-09-09": 8, "2026-09-10": 8,
			"2026-09-11": 8, "2026-09-14": 8, "2026-09-15": 8, "2026-09-16": 8,
			"2026-09-17": 8, "2026-09-18": 8, "2026-09-21": 8, "2026-09-22": 8,
			"2026-09-23": 8, "2026-09-24": 8, "2026-09-25": 8, "2026-09-28": 8,
			"2026-09-29": 8, "2026-09-30": 8,
		},
		"lisi": {
			"2026-09-01": 4,  // 不足 8h
			"2026-09-02": 16, // 超额但达标
			// 09-03 起未填
		},
		// wangwu：活跃任务指派人但整月无工时
	}
	assignees := map[string]bool{"zhangsan": true, "lisi": true, "wangwu": true, "zhaoliu": true}
	names := map[string]string{"zhangsan": "张三", "lisi": "李四", "wangwu": "王五"}
	// zhaoliu 在 activeAssignees 里但非优先、无数据 → 也会被列为无工时

	report := buildDailyCheckReport(now, 1029, "银河麒麟可观测平台V2.0", 8, periodStart, periodEnd,
		daily, assignees, names, "提醒", "", "月度工时检查", []string{"lisi"}, "https://zentao.kylin.me")

	if report.Workdays != 22 {
		t.Fatalf("应填工作日应为 22，得到 %d", report.Workdays)
	}
	if report.TotalPeople != 4 {
		t.Fatalf("检查人数应为 4，得到 %d", report.TotalPeople)
	}
	if report.OkCount != 1 || report.IssueCount != 3 {
		t.Fatalf("达标/未达标应为 1/3，得到 %d/%d", report.OkCount, report.IssueCount)
	}
	// lisi: 09-01(不足) + 09-02 起全部工作日缺失 = 1 + 21 = 22... 09-02 达标，缺失为 09-01 与 09-03~09-30 的 21 天 = 22 天中缺 21 天
	var lisi *models.AssigneeDailyCheckStats
	for i := range report.Details {
		if report.Details[i].Account == "lisi" {
			lisi = &report.Details[i]
		}
	}
	if lisi == nil {
		t.Fatal("报告中应包含 lisi")
	}
	// 09-01 填了 4h（不足）、09-02 填了 16h（达标），其余 20 个工作日未填
	if len(lisi.MissingDays) != 21 {
		t.Fatalf("lisi 应缺 21 个工作日，得到 %d", len(lisi.MissingDays))
	}
	if lisi.MissingDays[0].Date != "2026-09-01" || lisi.MissingDays[0].Hours != 4 {
		t.Fatalf("lisi 首个未达标日应为 09-01 4h，得到 %+v", lisi.MissingDays[0])
	}
	if lisi.NoEffort {
		t.Fatal("lisi 有工时记录，不应标记为无工时")
	}
	// lisi 21 + wangwu 22 + zhaoliu 22
	if report.TotalMissing != 65 {
		t.Fatalf("未达标人日应为 65，得到 %d", report.TotalMissing)
	}
	// 优先人员 lisi 应排最前
	if report.Details[0].Account != "lisi" {
		t.Fatalf("优先人员应置顶，得到 %s", report.Details[0].Account)
	}
	if !strings.Contains(report.Message, "周期内无任何工时记录") {
		t.Fatal("整月无工时的人员应在消息中标注")
	}
	if !strings.Contains(report.Message, "09-01 4.0h") {
		t.Fatal("填报不足的日期应展示已填小时数")
	}
	if !strings.Contains(report.Message, "等共 21 天") {
		t.Fatal("超过 10 条的缺失日期应折叠展示总数")
	}
	if !strings.Contains(report.Message, "https://zentao.kylin.me/timelog?") || !strings.Contains(report.Message, "product=1029") {
		t.Fatal("消息应附带 timelog 回访链接")
	}
}
