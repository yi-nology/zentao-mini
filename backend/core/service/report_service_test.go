package service

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/yi-nology/zentao-mini/backend/core/models"
)

func TestBuildViewURL(t *testing.T) {
	cases := []struct {
		name    string
		viewURL string
		path    string
		params  map[string]string
		want    string
	}{
		{"空回访地址返回空串", "", "/bugs", map[string]string{"product": "1029"}, ""},
		{"空白符与尾斜杠归一化", " https://zentao.kylin.me/ ", "/bugs", nil, "https://zentao.kylin.me/bugs"},
		{"带查询参数", "https://zentao.kylin.me", "/bugs", map[string]string{"product": "1029", "status": "active"}, "https://zentao.kylin.me/bugs?product=1029&status=active"},
		{"空值参数被忽略", "https://zentao.kylin.me", "/tasks", map[string]string{"product": "1", "project": ""}, "https://zentao.kylin.me/tasks?product=1"},
		{"参数值需转义", "https://zentao.kylin.me", "/bugs", map[string]string{"assignedTo": "a b"}, "https://zentao.kylin.me/bugs?assignedTo=a+b"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := buildViewURL(c.viewURL, c.path, c.params)
			if got != c.want {
				t.Fatalf("buildViewURL(%q, %q, %v) = %q, want %q", c.viewURL, c.path, c.params, got, c.want)
			}
		})
	}
}

func TestDetailLinkParams(t *testing.T) {
	cases := []struct {
		name         string
		productID    int
		projectID    int
		statusFilter string
		wantKeys     []string
	}{
		{"仅产品", 1029, 0, "active", []string{"product", "status"}},
		{"产品加项目", 1029, 7, "", []string{"product", "project"}},
		{"all 不过滤状态", 1, 0, "all", []string{"product"}},
		{"active-resolved 不过滤状态", 1, 0, "active-resolved", []string{"product"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			params := detailLinkParams(c.productID, c.projectID, c.statusFilter)
			if len(params) != len(c.wantKeys) {
				t.Fatalf("params = %v, 期望 %d 个键", params, len(c.wantKeys))
			}
			for _, k := range c.wantKeys {
				if _, ok := params[k]; !ok {
					t.Fatalf("params 缺少键 %q: %v", k, params)
				}
			}
		})
	}
}

func TestZentaoBugURL(t *testing.T) {
	if got := zentaoBugURL("https://pm.kylin.com/", 456); got != "https://pm.kylin.com/bug-view-456.html" {
		t.Fatalf("zentaoBugURL = %q", got)
	}
	if got := zentaoBugURL("", 456); got != "" {
		t.Fatalf("未配置禅道地址时应返回空串, got %q", got)
	}
}

func TestBuildMessageWithViewURL(t *testing.T) {
	details := []models.AssigneeBugStats{
		{Assignee: "张三", Account: "zhangsan", Total: 5, Fatal: 1, Serious: 1, Moderate: 2},
		{Assignee: "李四", Account: "lisi", Total: 2, Suggest: 2},
	}
	breakdown := map[string]int{"active": 7, "resolved": 1, "closed": 1}
	msg := buildMessage("Bug 分布报告 - 测试", time.Now(), 7, "active", details, breakdown, "提醒", "", "备注头", "https://zentao.kylin.me/bugs?product=1029&status=active")

	if !strings.Contains(msg, "🔗 查看详情：https://zentao.kylin.me/bugs?product=1029&status=active") {
		t.Fatalf("消息缺少查看详情链接:\n%s", msg)
	}
	if !strings.Contains(msg, "📊 活跃 Bug：7个，严重级别（致命+严重）2个") {
		t.Fatalf("汇总行口径不符合预期:\n%s", msg)
	}
	if !strings.Contains(msg, "👤 张三  5个，严重 2个\n   └ 致命:1 严重:1 一般:2") {
		t.Fatalf("指派人行应只显示非零档:\n%s", msg)
	}
	if !strings.Contains(msg, "👤 李四  2个\n   └ 建议:2") {
		t.Fatalf("纯建议级 Bug 不应显示严重后缀:\n%s", msg)
	}
	if !strings.Contains(msg, "⚠️ 严重级别共 2个（致命 1 | 严重 1）") || !strings.Contains(msg, "💡 另有建议级 2个") {
		t.Fatalf("尾部汇总行不符合预期:\n%s", msg)
	}
	if strings.Contains(msg, "高级别") {
		t.Fatalf("消息不应再出现\"高级别\"旧口径:\n%s", msg)
	}
}

func TestBuildMessageWithoutViewURL(t *testing.T) {
	msg := buildMessage("Bug 分布报告 - 测试", time.Now(), 0, "", nil, map[string]int{}, "提醒", "", "", "")
	if strings.Contains(msg, "查看详情") {
		t.Fatalf("未配置回访地址时不应出现链接行:\n%s", msg)
	}
	if !strings.Contains(msg, "📊 活跃 Bug：0个，严重级别（致命+严重）0个") {
		t.Fatalf("空数据汇总行不符合预期:\n%s", msg)
	}
	if strings.Contains(msg, "⚠️") {
		t.Fatalf("无严重级别 Bug 时不应出现警示行:\n%s", msg)
	}
}

func TestBuildRequirementMessageDeterministic(t *testing.T) {
	details := []models.AssigneeStoryStats{{Assignee: "张三", Total: 4, Active: 2, Closed: 1, Accepted: 1}}
	breakdown := map[string]int{"active": 2, "closed": 1, "draft": 3, "changed": 5}

	first := buildRequirementMessage("需求进度报告", time.Now(), 11, details, breakdown, "", "", "", "")
	for i := 0; i < 20; i++ {
		again := buildRequirementMessage("需求进度报告", time.Now(), 11, details, breakdown, "", "", "", "")
		if again != first {
			t.Fatalf("需求状态分布顺序不确定:\n第一次:\n%s\n第%d次:\n%s", first, i+2, again)
		}
	}

	want := "📈 需求状态分布：激活 2 | 已变更 5 | 已关闭 1 | 草稿 3"
	if !strings.Contains(first, want) {
		t.Fatalf("状态分布排序/中文标签不符合预期，期望包含 %q:\n%s", want, first)
	}
}

func TestBuildTaskMessageBreakdown(t *testing.T) {
	details := []models.TaskProgressStats{{Assignee: "张三", Total: 3, Done: 2, Doing: 1, Estimate: 10, Consumed: 5, Progress: 50}}
	breakdown := map[string]int{"wait": 1, "doing": 1, "done": 2, "cancel": 1}
	msg := buildTaskMessage("任务进度报告", time.Now(), 5, 10, 5, 50, details, breakdown, "", "", "", "https://zentao.kylin.me/tasks?product=1029")

	if !strings.Contains(msg, "🔗 查看详情：https://zentao.kylin.me/tasks?product=1029") {
		t.Fatalf("任务报告缺少查看详情链接:\n%s", msg)
	}
	if !strings.Contains(msg, "📈 任务状态分布：待开始 1 | 进行中 1 | 已完成 2 | 已取消 1") {
		t.Fatalf("任务状态分布不符合预期:\n%s", msg)
	}
}

func TestBuildBugAgingMessageTruncatesPerPerson(t *testing.T) {
	bugs := make([]models.BugAgingItem, 0, 8)
	for i := 1; i <= 8; i++ {
		bugs = append(bugs, models.BugAgingItem{ID: i, Title: fmt.Sprintf("缺陷%d", i), Severity: "一般", DaysOpen: 30 + i})
	}
	details := []models.AssigneeBugAgingStats{{Assignee: "张三", Total: 8, Bugs: bugs}}
	msg := buildBugAgingMessage("超时提醒", time.Now(), 8, 30, details, "", "", "", "", "")

	if got := strings.Count(msg, "   └ "); got != agingMaxListPerPerson+1 {
		t.Fatalf("每人最多列出 %d 条+1 条折叠提示, 实际 %d 条:\n%s", agingMaxListPerPerson, got, msg)
	}
	if !strings.Contains(msg, "└ …等 3 条未列出，详见上方链接") {
		t.Fatalf("超出部分应折叠提示:\n%s", msg)
	}
	if !strings.Contains(msg, "（阈值：30天，最长已停留 38天）") {
		t.Fatalf("汇总行应包含最长停留天数:\n%s", msg)
	}
}

func TestBuildBugAgingMessageWithLinks(t *testing.T) {
	details := []models.AssigneeBugAgingStats{
		{
			Assignee: "张三",
			Total:    1,
			Bugs: []models.BugAgingItem{
				{ID: 456, Title: "登录页面白屏", Severity: "严重", OpenedDate: "2026-09-01", DaysOpen: 12},
			},
		},
	}
	msg := buildBugAgingMessage("超时提醒", time.Now(), 1, 7, details, "提醒", "", "", "https://zentao.kylin.me/bugs?product=1029", "https://pm.kylin.com")

	if !strings.Contains(msg, "🔗 查看详情：https://zentao.kylin.me/bugs?product=1029") {
		t.Fatalf("超时提醒缺少查看详情链接:\n%s", msg)
	}
	if !strings.Contains(msg, "🔗 https://pm.kylin.com/bug-view-456.html") {
		t.Fatalf("超时 Bug 缺少禅道详情直链:\n%s", msg)
	}

	// 未配置禅道地址时回退为 #ID 形式
	noLink := buildBugAgingMessage("超时提醒", time.Now(), 1, 7, details, "提醒", "", "", "", "")
	if !strings.Contains(noLink, "└ #456 [严重]") || strings.Contains(noLink, "bug-view") {
		t.Fatalf("未配置禅道地址时应回退 #ID 格式:\n%s", noLink)
	}
}
