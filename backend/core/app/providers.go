package app

import (
	"os"

	"github.com/yi-nology/zentao-mini/backend/core/initialization"
	"github.com/yi-nology/zentao-mini/backend/core/zentao"
)

// 真实参与编译的 provider 实现（issue #30：此前 provider 副本藏在 wire_gen.go
// 里手工维护，与 wire.go 无限漂移；现在单源在本文件，wire.go 只留 injector 声明，
// wire_gen.go 由 wire 工具重新生成）。

func provideInitService(config *AppConfig) *initialization.InitService {
	return initialization.NewInitService(
		config.AuthDBPath,
		config.EncryptionKey,
	)
}

func provideConfigStore(config *AppConfig) *initialization.ConfigStore {
	return initialization.NewConfigStore(config.AuthDBPath)
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

	// 如果仍然没有配置，从环境变量获取（优先 ZENTAO_MINI_ 前缀）
	if zentaoServer == "" {
		zentaoServer = os.Getenv("ZENTAO_MINI_ZENTAO_SERVER")
		if zentaoServer == "" {
			zentaoServer = os.Getenv("ZENTAO_SERVER")
		}
	}
	if zentaoAccount == "" {
		zentaoAccount = os.Getenv("ZENTAO_MINI_ZENTAO_ACCOUNT")
		if zentaoAccount == "" {
			zentaoAccount = os.Getenv("ZENTAO_ACCOUNT")
		}
	}
	if zentaoPassword == "" {
		zentaoPassword = os.Getenv("ZENTAO_MINI_ZENTAO_PASSWORD")
		if zentaoPassword == "" {
			zentaoPassword = os.Getenv("ZENTAO_PASSWORD")
		}
	}

	// 配置面接上（issue #21/#30）：AppConfig 的两个字段已是 time.Duration
	//（main.go 从 yaml 换算：refresh 小时、timeout 秒），此处直接透传，
	// 不再二次乘 time.Second——那会把 12h 溢出成负数、又被 <=0 缺省分支吞掉。
	// 0 = 未配置，client 内部取缺省（12h / 120s）。
	return zentao.NewClientWithConfig(zentaoServer, zentaoAccount, zentaoPassword,
		config.ZentaoTokenRefreshInterval,
		config.ZentaoRequestTimeout)
}

// provideHTTPApp 提供HTTPApp并绑定到Application接口
func provideHTTPApp(config *AppConfig, deps *Dependencies) Application {
	return NewHTTPApp(config, deps)
}
