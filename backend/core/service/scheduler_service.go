package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"chandao-mini/backend/core/initialization"
	"chandao-mini/backend/core/logger"
	"chandao-mini/backend/core/models"

	"go.uber.org/zap"
)

type SchedulerService struct {
	store    *initialization.ConfigStore
	cron     *cron.Cron
	report   *ReportService
	webhook  *WebhookService
	entryMap map[string]cron.EntryID
	mu       sync.Mutex
}

func NewSchedulerService(
	store *initialization.ConfigStore,
	reportService *ReportService,
	webhookService *WebhookService,
) *SchedulerService {
	return &SchedulerService{
		store:    store,
		cron:     cron.New(cron.WithSeconds()),
		report:   reportService,
		webhook:  webhookService,
		entryMap: make(map[string]cron.EntryID),
	}
}

func (s *SchedulerService) Start() error {
	tasks, err := s.store.ListTasks()
	if err != nil {
		return fmt.Errorf("加载定时任务失败: %w", err)
	}
	for i := range tasks {
		if tasks[i].Enabled {
			if err := s.registerTask(&tasks[i]); err != nil {
				logger.Error("注册定时任务失败",
					zap.String("taskID", tasks[i].ID),
					zap.String("taskName", tasks[i].Name),
					zap.Error(err))
			}
		}
	}
	s.cron.Start()
	logger.Info("定时任务调度器已启动", zap.Int("taskCount", len(tasks)))
	return nil
}

func (s *SchedulerService) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.Info("定时任务调度器已停止")
}

func (s *SchedulerService) registerTask(task *models.SchedulerTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if oldID, ok := s.entryMap[task.ID]; ok {
		s.cron.Remove(oldID)
		delete(s.entryMap, task.ID)
	}

	cronExpr := task.CronExpr
	if len(cronExpr) == 5 {
		cronExpr = "0 " + cronExpr
	}

	taskCopy := *task
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeTask(&taskCopy)
	})
	if err != nil {
		return fmt.Errorf("解析cron表达式失败 '%s': %w", task.CronExpr, err)
	}

	s.entryMap[task.ID] = entryID
	logger.Info("定时任务已注册",
		zap.String("taskID", task.ID),
		zap.String("taskName", task.Name),
		zap.String("cronExpr", task.CronExpr))
	return nil
}

func (s *SchedulerService) unregisterTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entryID, ok := s.entryMap[taskID]; ok {
		s.cron.Remove(entryID)
		delete(s.entryMap, taskID)
	}
}

func (s *SchedulerService) ListTasks() ([]models.SchedulerTask, error) {
	return s.store.ListTasks()
}

func (s *SchedulerService) GetTask(id string) (*models.SchedulerTask, error) {
	return s.store.GetTask(id)
}

func (s *SchedulerService) CreateTask(task *models.SchedulerTask) error {
	now := time.Now()
	task.ID = uuid.New().String()
	task.CreatedAt = now
	task.UpdatedAt = now
	if task.StatusFilter == "" {
		task.StatusFilter = "active"
	}
	for i := range task.Webhooks {
		if task.Webhooks[i].ID == "" {
			task.Webhooks[i].ID = uuid.New().String()
		}
	}
	if err := s.store.CreateTask(task); err != nil {
		return err
	}
	if task.Enabled {
		return s.registerTask(task)
	}
	return nil
}

func (s *SchedulerService) UpdateTask(task *models.SchedulerTask) error {
	task.UpdatedAt = time.Now()
	for i := range task.Webhooks {
		if task.Webhooks[i].ID == "" {
			task.Webhooks[i].ID = uuid.New().String()
		}
	}
	if err := s.store.UpdateTask(task); err != nil {
		return err
	}
	s.unregisterTask(task.ID)
	if task.Enabled {
		return s.registerTask(task)
	}
	return nil
}

func (s *SchedulerService) DeleteTask(id string) error {
	s.unregisterTask(id)
	return s.store.DeleteTask(id)
}

func (s *SchedulerService) ToggleTask(id string) (*models.SchedulerTask, error) {
	task, err := s.store.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("任务不存在")
	}
	task.Enabled = !task.Enabled
	task.UpdatedAt = time.Now()
	if err := s.store.UpdateTask(task); err != nil {
		return nil, err
	}
	s.unregisterTask(id)
	if task.Enabled {
		if err := s.registerTask(task); err != nil {
			return nil, err
		}
	}
	return task, nil
}

func (s *SchedulerService) RunTaskNow(id string) (*models.TaskExecutionLog, error) {
	task, err := s.store.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("任务不存在")
	}
	return s.executeTask(task), nil
}

func (s *SchedulerService) executeTask(task *models.SchedulerTask) *models.TaskExecutionLog {
	logEntry := &models.TaskExecutionLog{
		ID:        uuid.New().String(),
		TaskID:    task.ID,
		TaskName:  task.Name,
		StartedAt: time.Now(),
		Status:    "running",
	}

	logger.Info("开始执行定时任务",
		zap.String("taskID", task.ID),
		zap.String("taskName", task.Name))

	report, err := s.report.GenerateBugReport(task.ProjectID, task.ProjectName, task.StatusFilter, task.Keyword)
	if err != nil {
		logEntry.Status = "failed"
		logEntry.Error = err.Error()
		now := time.Now()
		logEntry.FinishedAt = &now
		_ = s.store.SaveLog(logEntry)
		logger.Error("定时任务执行失败", zap.String("taskID", task.ID), zap.Error(err))
		return logEntry
	}

	logEntry.BugTotal = report.Total
	logEntry.HighSeverity = report.HighSeverity
	logEntry.AssigneeCount = len(report.Details)

	webhookResults := s.webhook.SendAll(task.Webhooks, report)
	logEntry.WebhookResults = webhookResults

	successCount := 0
	for _, r := range webhookResults {
		if r.Success {
			successCount++
		}
	}

	enabledCount := 0
	for _, wh := range task.Webhooks {
		if wh.Enabled {
			enabledCount++
		}
	}

	now := time.Now()
	logEntry.FinishedAt = &now

	if enabledCount == 0 {
		logEntry.Status = "success"
	} else if successCount == 0 {
		logEntry.Status = "failed"
	} else if successCount < enabledCount {
		logEntry.Status = "partial"
	} else {
		logEntry.Status = "success"
	}

	_ = s.store.SaveLog(logEntry)
	logger.Info("定时任务执行完成",
		zap.String("taskID", task.ID),
		zap.String("status", logEntry.Status),
		zap.Int("bugTotal", report.Total))
	return logEntry
}

func (s *SchedulerService) ListLogs(taskID string, limit int) ([]models.TaskExecutionLog, error) {
	return s.store.ListLogs(taskID, limit)
}
