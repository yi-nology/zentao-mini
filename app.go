package main

import (
	"context"
	"log"

	"github.com/yi-nology/zentao-mini/backend/core/app"
	"github.com/yi-nology/zentao-mini/backend/core/event"
	"github.com/yi-nology/zentao-mini/backend/core/logger"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/zap"
)

// App struct - Wails应用包装器
// 该结构体包装了core/app包中的WailsApp，提供Wails框架所需的接口
type App struct {
	ctx       context.Context
	wailsApp  *app.WailsApp
	notifUnsub func()
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	config := &app.AppConfig{
		Type:          "wails",
		Port:          "",
		AuthDBPath:    "",
		EncryptionKey: "",
	}

	application, err := app.InitializeWailsApp(config)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	wailsApp, ok := application.(*app.WailsApp)
	if !ok {
		log.Fatalf("Expected WailsApp, got %T", application)
	}

	a.wailsApp = wailsApp

	if err := wailsApp.Start(ctx); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// 订阅事件总线，把后端事件桥接到前端
	a.subscribeNotifications()
}

// subscribeNotifications 订阅定时任务事件，转发到前端
// 前端通过 EventsOn('notification', ...) 接收，触发桌面通知
func (a *App) subscribeNotifications() {
	bus := event.GetGlobalBus()
	a.notifUnsub = bus.Subscribe(event.TaskCompleted, func(e event.Event) {
		payload, ok := e.Payload.(map[string]interface{})
		if !ok {
			return
		}
		taskName, _ := payload["taskName"].(string)
		status, _ := payload["status"].(string)

		// 通过 wails runtime 把事件推到前端
		runtime.EventsEmit(a.ctx, "notification", map[string]interface{}{
			"type:":     "task",
			"title":     "定时任务执行完成",
			"body":      "任务「" + taskName + "」状态: " + status,
			"taskId":    payload["taskID"],
			"taskName":  taskName,
			"status":    status,
			"bugTotal":  payload["bugTotal"],
		})
		logger.Info("notification emitted",
			zap.String("taskName", taskName),
			zap.String("status", status))
	})

	bus.Subscribe(event.TaskFailed, func(e event.Event) {
		payload, _ := e.Payload.(map[string]interface{})
		taskName, _ := payload["taskName"].(string)
		runtime.EventsEmit(a.ctx, "notification", map[string]interface{}{
			"type":     "task",
			"level":    "error",
			"title":    "定时任务执行失败",
			"body":     "任务「" + taskName + "」执行失败",
			"taskName": taskName,
		})
	})
}

// ShowWindow 显示主窗口（供前端从托盘恢复时调用）
func (a *App) ShowWindow() {
	if a.ctx == nil {
		return
	}
	runtime.WindowShow(a.ctx)
}

// HideWindow 隐藏主窗口到托盘（供前端调用）
func (a *App) HideWindow() {
	if a.ctx == nil {
		return
	}
	runtime.WindowHide(a.ctx)
}

// MinimizeToTray 最小化到托盘（前端调用入口）
func (a *App) MinimizeToTray() {
	a.HideWindow()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	if a.wailsApp != nil {
		return a.wailsApp.Greet(name)
	}
	return "Hello " + name
}

// shutdown is called when the app exits
func (a *App) shutdown(ctx context.Context) {
	if a.notifUnsub != nil {
		a.notifUnsub()
	}
	if a.wailsApp != nil {
		if err := a.wailsApp.Stop(ctx); err != nil {
			log.Printf("Failed to stop application: %v", err)
		}
	}
}
