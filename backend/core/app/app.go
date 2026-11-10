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
	// ZentaoTokenRefreshInterval token 刷新间隔（已是 Duration；0 = client 缺省 12h）。
	// main.go 从 Config.Zentao.TokenRefreshInterval（小时）换算，换算只在那一处。
	ZentaoTokenRefreshInterval time.Duration
	// ZentaoRequestTimeout 请求超时（已是 Duration；0 = client 缺省 120s）。
	// main.go 从 Config.Zentao.RequestTimeout（秒）换算，换算只在那一处。
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
