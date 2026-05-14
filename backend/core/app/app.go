package app

import (
	"context"

	"chandao-mini/backend/core/handlers"
	"chandao-mini/backend/core/initialization"
	"chandao-mini/backend/core/zentao"
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

	AuthDBPath    string
	EncryptionKey string

	StaticPath string
}

type Dependencies struct {
	InitService  *initialization.InitService
	ZentaoClient *zentao.Client
	ConfigStore  *initialization.ConfigStore
	Handlers     *handlers.HandlerRegistry
}

func NewDependencies(initService *initialization.InitService, zentaoClient *zentao.Client, configStore *initialization.ConfigStore) *Dependencies {
	deps := &Dependencies{
		InitService:  initService,
		ZentaoClient: zentaoClient,
		ConfigStore:  configStore,
		Handlers:     handlers.NewHandlerRegistry(zentaoClient, initService),
	}
	deps.Handlers.InitScheduler(configStore)
	return deps
}
