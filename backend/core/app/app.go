package app

import (
	"context"

	"chandao-mini/backend/core/handlers"
	"chandao-mini/backend/core/initialization"
	"chandao-mini/backend/core/zentao"
)

// Application 定义应用程序的统一接口
// 该接口抽象了不同运行模式（Wails桌面应用、HTTP服务器）的共同行为
type Application interface {
	// Start 启动应用程序
	// ctx 提供应用程序的生命周期管理
	Start(ctx context.Context) error

	// Stop 停止应用程序，执行优雅关闭
	Stop(ctx context.Context) error

	// Name 返回应用程序名称
	Name() string
}

// AppConfig 应用程序配置
type AppConfig struct {
	// 应用类型: "wails" 或 "http"
	Type string

	// HTTP服务器配置
	Port string

	// 禅道配置
	ZentaoServer   string
	ZentaoAccount  string
	ZentaoPassword string

	// 认证配置
	AuthConfigPath string
	AuthDBPath     string
	EncryptionKey  string

	// 静态资源路径（仅用于HTTP模式）
	StaticPath string
}

// Dependencies 应用程序依赖项
// 使用依赖注入模式，确保所有组件只初始化一次
type Dependencies struct {
	// 核心服务
	InitService  *initialization.InitService
	ZentaoClient *zentao.Client

	// Handler注册表 - 所有handler只初始化一次
	Handlers *handlers.HandlerRegistry
}

// NewDependencies 创建依赖项实例
func NewDependencies(initService *initialization.InitService, zentaoClient *zentao.Client) *Dependencies {
	return &Dependencies{
		InitService:  initService,
		ZentaoClient: zentaoClient,
		Handlers:     handlers.NewHandlerRegistry(zentaoClient),
	}
}
