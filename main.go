package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	// Off-white background color (#FAF9F6)
	bgColor := &options.RGBA{R: 250, G: 249, B: 246, A: 1}

	err := wails.Run(&options.App{
		Title:  "Dad's PDF Tools",
		Width:  900,
		Height: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: bgColor,
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			Theme: windows.Light,
			CustomTheme: &windows.ThemeSettings{
				// Title bar colors (active window)
				DarkModeTitleBar:   windows.RGB(250, 249, 246),
				DarkModeTitleText:  windows.RGB(26, 26, 26),
				LightModeTitleBar:  windows.RGB(250, 249, 246),
				LightModeTitleText: windows.RGB(26, 26, 26),
				// Title bar colors (inactive window)
				DarkModeTitleBarInactive:   windows.RGB(245, 244, 239),
				DarkModeTitleTextInactive:  windows.RGB(102, 102, 102),
				LightModeTitleBarInactive:  windows.RGB(245, 244, 239),
				LightModeTitleTextInactive: windows.RGB(102, 102, 102),
				// Border
				DarkModeBorder:  windows.RGB(229, 227, 219),
				LightModeBorder: windows.RGB(229, 227, 219),
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
