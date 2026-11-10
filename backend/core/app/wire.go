//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
)

// InitializeHTTPApp 初始化HTTP应用
// 使用wire自动生成依赖注入代码（make generate 或 go generate ./backend/core/app/）
// provider 实现在 providers.go（单源）；本文件只做 injector 声明，不参与编译。
func InitializeHTTPApp(config *AppConfig) (Application, error) {
	wire.Build(
		provideInitService,
		provideZentaoClient,
		provideConfigStore,
		NewDependencies,
		provideHTTPApp,
	)
	return nil, nil
}

// InitializeWailsApp 已随桌面模式移除（v1.4.0），唯一入口 InitializeHTTPApp。
