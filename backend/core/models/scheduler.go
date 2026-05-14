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
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Enabled      bool            `json:"enabled"`
	CronExpr     string          `json:"cronExpr"`
	Webhooks     []WebhookConfig `json:"webhooks"`
	ProjectID    int             `json:"projectId"`
	ProductID    int             `json:"productId"`
	ProjectName  string          `json:"projectName"`
	ProductName  string          `json:"productName"`
	StatusFilter string          `json:"statusFilter"`
	Keyword      string          `json:"keyword"`
	LastRunAt    *time.Time      `json:"lastRunAt"`
	LastRunStatus string         `json:"lastRunStatus"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

type TaskExecutionLog struct {
	ID             string           `json:"id"`
	TaskID         string           `json:"taskId"`
	TaskName       string           `json:"taskName"`
	StartedAt      time.Time        `json:"startedAt"`
	FinishedAt     *time.Time       `json:"finishedAt"`
	Status         string           `json:"status"`
	BugTotal       int              `json:"bugTotal"`
	HighSeverity   int              `json:"highSeverity"`
	AssigneeCount  int              `json:"assigneeCount"`
	WebhookResults []WebhookResult  `json:"webhookResults"`
	Error          string           `json:"error,omitempty"`
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
}

type BugReport struct {
	Title           string                    `json:"title"`
	Timestamp       string                    `json:"timestamp"`
	ProjectName     string                    `json:"projectName"`
	Total           int                       `json:"total"`
	HighSeverity    int                       `json:"highSeverity"`
	StatusBreakdown map[string]int            `json:"statusBreakdown"`
	Details         []AssigneeBugStats        `json:"details"`
	Message         string                    `json:"message"`
}

type CronDB struct {
	Tasks []SchedulerTask    `json:"tasks"`
	Logs  []TaskExecutionLog `json:"logs"`
}
