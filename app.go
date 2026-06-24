package main

import (
	"context"
	"log"

	"github.com/yi-nology/zentao-mini/backend/core/app"

	"github.com/joho/godotenv"
)

// App struct - Wails应用包装器
// 该结构体包装了core/app包中的WailsApp，提供Wails框架所需的接口
type App struct {
	wailsApp *app.WailsApp
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
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
	if a.wailsApp != nil {
		if err := a.wailsApp.Stop(ctx); err != nil {
			log.Printf("Failed to stop application: %v", err)
		}
	}
}
