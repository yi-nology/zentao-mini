package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "zentao-mini",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1 },
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		// 拦截窗口关闭：默认改为隐藏到托盘，避免误关后无法恢复
		// 用户必须通过托盘菜单"退出"才能真正退出
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			wailsruntime.WindowHide(ctx)
			// 同时发个事件通知前端（如果前端要响应）
			wailsruntime.EventsEmit(ctx, "window:hidden")
			return true // 阻止真正关闭
		},
		Bind: []interface{}{
			app,
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
		Mac: &mac.Options{
			WebviewIsTransparent: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
