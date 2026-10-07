package models

import "time"

type WebhookConfig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Enabled  bool   `json:"enabled"`
	Platform string `json:"platform"` // generic | lanxin
	Secret   string `json:"secret"`   // 蓝信加签密钥
	SkipSSL  bool   `json:"skipSSL"`  // 跳过SSL证书验证
}

type SchedulerTask struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Enabled           bool            `json:"enabled"`
	CronExpr          string          `json:"cronExpr"`
	Webhooks          []WebhookConfig `json:"webhooks"`
	ProjectID         int             `json:"projectId"`
	ProductID         int             `json:"productId"`
	ProjectName       string          `json:"projectName"`
	ProductName       string          `json:"productName"`
	StatusFilter      string          `json:"statusFilter"`
	ReportType        string          `json:"reportType"`        // bug | requirement | task | bug-aging | daily-report-check
	AgingDays         int             `json:"agingDays"`         // bug-aging 超时天数阈值，默认 7
	CheckHours        float64         `json:"checkHours"`        // daily-report-check 每工作日最低工时阈值，默认 8
	PriorityAssignees []string        `json:"priorityAssignees"` // 优先展示的人员账号列表
	MessageHeader     string          `json:"messageHeader"`     // 消息头备注，如"详情查看 xxxx"
	Keyword           string          `json:"keyword"`
	ExternalInfo      string          `json:"externalInfo"`
	ViewURL           string          `json:"viewURL"` // 回访地址（如 https://zentao.kylin.me），推送时自动附加"查看详情"深链
	LastRunAt         *time.Time      `json:"lastRunAt"`
	LastRunStatus     string          `json:"lastRunStatus"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

type TaskExecutionLog struct {
	ID             string          `json:"id"`
	TaskID         string          `json:"taskId"`
	TaskName       string          `json:"taskName"`
	StartedAt      time.Time       `json:"startedAt"`
	FinishedAt     *time.Time      `json:"finishedAt"`
	Status         string          `json:"status"`
	BugTotal       int             `json:"bugTotal"`
	HighSeverity   int             `json:"highSeverity"`
	AssigneeCount  int             `json:"assigneeCount"`
	WebhookResults []WebhookResult `json:"webhookResults"`
	Error          string          `json:"error,omitempty"`
}

type WebhookResult struct {
	WebhookID   string `json:"webhookId"`
	WebhookName string `json:"webhookName"`
	WebhookURL  string `json:"webhookUrl"`
	Success     bool   `json:"success"`
	StatusCode  int    `json:"statusCode,omitempty"`
	Error       string `json:"error,omitempty"`
}

type AssigneeBugStats struct {
	Assignee     string `json:"assignee"`
	Account      string `json:"account"`
	Total        int    `json:"total"`
	HighSeverity int    `json:"highSeverity"`
	Fatal        int    `json:"fatal"`
	Serious      int    `json:"serious"`
	Moderate     int    `json:"moderate"`
	Minor        int    `json:"minor"`
	Suggest      int    `json:"suggest"` // 5级=建议（禅道自定义第五档）
}

type BugReport struct {
	Title           string             `json:"title"`
	Timestamp       string             `json:"timestamp"`
	ProjectName     string             `json:"projectName"`
	Total           int                `json:"total"`
	HighSeverity    int                `json:"highSeverity"`
	StatusBreakdown map[string]int     `json:"statusBreakdown"`
	Details         []AssigneeBugStats `json:"details"`
	Message         string             `json:"message"`
}

type CronDB struct {
	Tasks []SchedulerTask    `json:"tasks"`
	Logs  []TaskExecutionLog `json:"logs"`
}

type AssigneeStoryStats struct {
	Assignee  string `json:"assignee"`
	Account   string `json:"account"`
	Total     int    `json:"total"`
	Active    int    `json:"active"`
	Changed   int    `json:"changed"`
	Closed    int    `json:"closed"`
	Resolved  int    `json:"resolved"`
	Accepted  int    `json:"accepted"`
	Reviewing int    `json:"reviewing"`
}

type RequirementReport struct {
	Title           string               `json:"title"`
	Timestamp       string               `json:"timestamp"`
	ProjectName     string               `json:"projectName"`
	ProductName     string               `json:"productName"`
	Total           int                  `json:"total"`
	StatusBreakdown map[string]int       `json:"statusBreakdown"`
	Details         []AssigneeStoryStats `json:"details"`
	Message         string               `json:"message"`
}

type TaskProgressStats struct {
	Assignee  string  `json:"assignee"`
	Account   string  `json:"account"`
	Total     int     `json:"total"`
	Wait      int     `json:"wait"`
	Doing     int     `json:"doing"`
	Done      int     `json:"done"`
	Paused    int     `json:"paused"`
	Cancelled int     `json:"cancelled"`
	Estimate  float64 `json:"estimate"`
	Consumed  float64 `json:"consumed"`
	Progress  float64 `json:"progress"`
}

type TaskProgressReport struct {
	Title           string              `json:"title"`
	Timestamp       string              `json:"timestamp"`
	ProjectName     string              `json:"projectName"`
	ProductName     string              `json:"productName"`
	Total           int                 `json:"total"`
	StatusBreakdown map[string]int      `json:"statusBreakdown"`
	TotalEstimate   float64             `json:"totalEstimate"`
	TotalConsumed   float64             `json:"totalConsumed"`
	OverallProgress float64             `json:"overallProgress"`
	Details         []TaskProgressStats `json:"details"`
	Message         string              `json:"message"`
}

type BugAgingItem struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Severity   string `json:"severity"`
	OpenedDate string `json:"openedDate"`
	DaysOpen   int    `json:"daysOpen"`
}

type AssigneeBugAgingStats struct {
	Assignee string         `json:"assignee"`
	Account  string         `json:"account"`
	Total    int            `json:"total"`
	Bugs     []BugAgingItem `json:"bugs"`
}

type BugAgingReport struct {
	Title       string                  `json:"title"`
	Timestamp   string                  `json:"timestamp"`
	ProjectName string                  `json:"projectName"`
	Total       int                     `json:"total"`
	AgingDays   int                     `json:"agingDays"`
	Details     []AssigneeBugAgingStats `json:"details"`
	Message     string                  `json:"message"`
}

// ========== 日报完成度检查（daily-report-check）==========

// DailyMissingDay 单个未达标工作日：Hours=0 表示当日未填报，>0 表示填报不足阈值
type DailyMissingDay struct {
	Date  string  `json:"date"`
	Hours float64 `json:"hours"`
}

type AssigneeDailyCheckStats struct {
	Assignee    string            `json:"assignee"`
	Account     string            `json:"account"`
	Workdays    int               `json:"workdays"`    // 应填报的工作日数
	OkDays      int               `json:"okDays"`      // 达标天数
	MissingDays []DailyMissingDay `json:"missingDays"` // 未达标日期明细
	TotalHours  float64           `json:"totalHours"`  // 周期内总工时
	NoEffort    bool              `json:"noEffort"`    // 周期内无任何工时记录（整月未填报）
}

type DailyReportCheckReport struct {
	Title        string                    `json:"title"`
	Timestamp    string                    `json:"timestamp"`
	ProductName  string                    `json:"productName"`
	PeriodStart  string                    `json:"periodStart"`
	PeriodEnd    string                    `json:"periodEnd"`
	Workdays     int                       `json:"workdays"`
	CheckHours   float64                   `json:"checkHours"`
	TotalPeople  int                       `json:"totalPeople"`
	OkCount      int                       `json:"okCount"`
	IssueCount   int                       `json:"issueCount"`
	TotalMissing int                       `json:"totalMissing"` // 未达标人日总数
	Details      []AssigneeDailyCheckStats `json:"details"`
	Message      string                    `json:"message"`
}
