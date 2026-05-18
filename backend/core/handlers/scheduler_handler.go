package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/yi-nology/zentao-mini/backend/core/errors"
	"github.com/yi-nology/zentao-mini/backend/core/models"
	"github.com/yi-nology/zentao-mini/backend/core/service"
)

type SchedulerHandler struct {
	schedulerService *service.SchedulerService
	webhookService   *service.WebhookService
}

func NewSchedulerHandler(schedulerService *service.SchedulerService, webhookService *service.WebhookService) *SchedulerHandler {
	return &SchedulerHandler{
		schedulerService: schedulerService,
		webhookService:   webhookService,
	}
}

func (h *SchedulerHandler) ListTasks(c *gin.Context) {
	tasks, err := h.schedulerService.ListTasks()
	if err != nil {
		errors.Error(c, errors.ExternalError("调度器", err))
		return
	}
	errors.Success(c, tasks)
}

func (h *SchedulerHandler) CreateTask(c *gin.Context) {
	var task models.SchedulerTask
	if err := c.ShouldBindJSON(&task); err != nil {
		errors.BadRequest(c, "请求参数格式错误")
		return
	}
	if task.Name == "" {
		errors.MissingParam(c, "name")
		return
	}
	if task.CronExpr == "" {
		errors.MissingParam(c, "cronExpr")
		return
	}
	if task.ProjectID == 0 {
		errors.MissingParam(c, "projectId")
		return
	}
	if len(task.Webhooks) == 0 {
		errors.MissingParam(c, "webhooks")
		return
	}
	if err := h.schedulerService.CreateTask(&task); err != nil {
		errors.Error(c, errors.NewInternalError("创建定时任务失败", err))
		return
	}
	errors.Success(c, task)
}

func (h *SchedulerHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.MissingParam(c, "id")
		return
	}
	var task models.SchedulerTask
	if err := c.ShouldBindJSON(&task); err != nil {
		errors.BadRequest(c, "请求参数格式错误")
		return
	}
	task.ID = id
	if err := h.schedulerService.UpdateTask(&task); err != nil {
		errors.Error(c, errors.NewInternalError("更新定时任务失败", err))
		return
	}
	errors.Success(c, task)
}

func (h *SchedulerHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.MissingParam(c, "id")
		return
	}
	if err := h.schedulerService.DeleteTask(id); err != nil {
		errors.Error(c, errors.NewInternalError("删除定时任务失败", err))
		return
	}
	errors.Success(c, nil)
}

func (h *SchedulerHandler) ToggleTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.MissingParam(c, "id")
		return
	}
	task, err := h.schedulerService.ToggleTask(id)
	if err != nil {
		errors.Error(c, errors.NewInternalError("切换定时任务状态失败", err))
		return
	}
	errors.Success(c, task)
}

func (h *SchedulerHandler) RunTaskNow(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.MissingParam(c, "id")
		return
	}
	logEntry, err := h.schedulerService.RunTaskNow(id)
	if err != nil {
		errors.Error(c, errors.NewInternalError("手动执行任务失败", err))
		return
	}
	errors.Success(c, logEntry)
}

func (h *SchedulerHandler) TestWebhook(c *gin.Context) {
	var req struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "请求参数格式错误")
		return
	}
	if req.URL == "" {
		errors.MissingParam(c, "url")
		return
	}
	result, err := h.webhookService.TestWebhook(req.URL)
	if err != nil {
		errors.Error(c, errors.NewInternalError("测试Webhook失败", err))
		return
	}
	errors.Success(c, result)
}

func (h *SchedulerHandler) GetTaskLogs(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.MissingParam(c, "id")
		return
	}
	logs, err := h.schedulerService.ListLogs(id, 50)
	if err != nil {
		errors.Error(c, errors.NewInternalError("获取执行日志失败", err))
		return
	}
	errors.Success(c, logs)
}

func (h *SchedulerHandler) GetAllLogs(c *gin.Context) {
	logs, err := h.schedulerService.ListLogs("", 100)
	if err != nil {
		errors.Error(c, errors.NewInternalError("获取执行日志失败", err))
		return
	}
	errors.Success(c, logs)
}
