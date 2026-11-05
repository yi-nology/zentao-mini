package app

import (
	"time"

	"context"
	"net/http"

	"github.com/yi-nology/zentao-mini/backend/core/handlers"
	"github.com/yi-nology/zentao-mini/backend/core/initialization"
	"github.com/yi-nology/zentao-mini/backend/core/zentao"
)

type Application interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

type AppConfig struct {
	Type string
	Port string

	ZentaoServer   string
	ZentaoAccount  string
	ZentaoPassword string
	// ZentaoTokenRefreshInterval token 刷新间隔（秒；0 = client 缺省 12h）。
	// 从 Config.Zentao.TokenRefreshInterval（小时）×3600 传入。
	ZentaoTokenRefreshInterval time.Duration
	// ZentaoRequestTimeout 请求超时（秒；0 = client 缺省 120s）。
	ZentaoRequestTimeout time.Duration

	AuthDBPath    string
	EncryptionKey string

	StaticPath string
	StaticFS   http.FileSystem
}

type Dependencies struct {
	InitService  *initialization.InitService
	ZentaoClient *zentao.Client
	ConfigStore  *initialization.ConfigStore
	Handlers     *handlers.HandlerRegistry
}

func NewDependencies(initService *initialization.InitService, zentaoClient *zentao.Client, configStore *initialization.ConfigStore) *Dependencies {
	return &Dependencies{
		InitService:  initService,
		ZentaoClient: zentaoClient,
		ConfigStore:  configStore,
		Handlers:     handlers.NewHandlerRegistry(zentaoClient, initService),
	}
}
