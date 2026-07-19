package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
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

	// 应用菜单：包含常用的窗口操作（macOS 顶部菜单 / Windows 窗口菜单）
	// 注意：菜单加速键仅在窗口获得焦点时生效，不是真正的全局快捷键
	// menu callback 没有 ctx，因此用闭包持有 app.ctx（startup 时已保存）
	appMenu := menu.NewMenuFromItems(
		menu.AppMenu(),
		menu.EditMenu(),
		menu.WindowMenu(),
		&menu.MenuItem{
			Label: "操作",
			SubMenu: &menu.Menu{
				Items: []*menu.MenuItem{
					{
						Label:      "最小化到托盘",
						Accelerator: keys.CmdOrCtrl("h"),
						Click: func(_ *menu.CallbackData) {
							if app.ctx != nil {
								wailsruntime.WindowHide(app.ctx)
							}
						},
					},
					{
						Label: "显示主窗口",
						Click: func(_ *menu.CallbackData) {
							if app.ctx != nil {
								wailsruntime.WindowShow(app.ctx)
							}
						},
					},
					{
						Label:      "刷新页面",
						Accelerator: keys.CmdOrCtrl("r"),
						Click: func(_ *menu.CallbackData) {
							if app.ctx != nil {
								wailsruntime.WindowReload(app.ctx)
							}
						},
					},
				},
			},
		},
	)

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
		Menu:             appMenu,
		// 拦截窗口关闭：默认改为隐藏到托盘，避免误关后无法恢复
		// 用户必须通过菜单"显示主窗口"恢复，或通过 OS 任务栏/dock 图标重新激活
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			wailsruntime.WindowHide(ctx)
			wailsruntime.EventsEmit(ctx, "window:hidden")
			return true
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
