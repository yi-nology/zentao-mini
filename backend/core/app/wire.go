//go:build wireinject
// +build wireinject

package app

import (
	"os"

	"chandao-mini/backend/core/initialization"
	"chandao-mini/backend/core/zentao"

	"github.com/google/wire"
)

// InitializeHTTPApp 初始化HTTP应用
// 使用wire自动生成依赖注入代码
func InitializeHTTPApp(config *AppConfig) (Application, error) {
	wire.Build(
		// 提供InitService
		provideInitService,

		// 提供ZentaoClient
		provideZentaoClient,

		// 提供Dependencies
		NewDependencies,

		// 提供HTTPApp并绑定到Application接口
		provideHTTPApp,
	)
	return nil, nil
}

// InitializeWailsApp 初始化Wails应用
// 使用wire自动生成依赖注入代码
func InitializeWailsApp(config *AppConfig) (Application, error) {
	wire.Build(
		// 提供InitService
		provideInitService,

		// 提供ZentaoClient
		provideZentaoClient,

		// 提供Dependencies
		NewDependencies,

		// 提供WailsApp并绑定到Application接口
		provideWailsApp,
	)
	return nil, nil
}

// provideInitService 提供InitService实例
func provideInitService(config *AppConfig) *initialization.InitService {
	return initialization.NewInitService(
		config.AuthConfigPath,
		config.AuthDBPath,
		config.EncryptionKey,
	)
}

// provideZentaoClient 提供ZentaoClient实例
func provideZentaoClient(config *AppConfig, initService *initialization.InitService) *zentao.Client {
	// 加载禅道配置
	zentaoServer, zentaoAccount, zentaoPassword := initialization.LoadZentaoConfig(initService)

	// 如果配置中有值，优先使用配置中的值
	if config.ZentaoServer != "" {
		zentaoServer = config.ZentaoServer
	}
	if config.ZentaoAccount != "" {
		zentaoAccount = config.ZentaoAccount
	}
	if config.ZentaoPassword != "" {
		zentaoPassword = config.ZentaoPassword
	}

	// 如果仍然没有配置，从环境变量获取
	if zentaoServer == "" {
		zentaoServer = os.Getenv("ZENTAO_SERVER")
	}
	if zentaoAccount == "" {
		zentaoAccount = os.Getenv("ZENTAO_ACCOUNT")
	}
	if zentaoPassword == "" {
		zentaoPassword = os.Getenv("ZENTAO_PASSWORD")
	}

	return zentao.NewClient(zentaoServer, zentaoAccount, zentaoPassword)
}

// provideHTTPApp 提供HTTPApp并绑定到Application接口
func provideHTTPApp(config *AppConfig, deps *Dependencies) Application {
	return NewHTTPApp(config, deps)
}

// provideWailsApp 提供WailsApp并绑定到Application接口
func provideWailsApp(config *AppConfig, deps *Dependencies) Application {
	return NewWailsApp(config, deps)
}
