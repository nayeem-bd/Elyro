package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"elyro/internal/store"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	dbPath, err := store.DefaultPath()
	if err != nil {
		log.Fatal(err)
	}
	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open database %s: %v", dbPath, err)
	}
	defer db.Close()

	app := application.New(application.Options{
		Name:        "Elyro",
		Description: "Offline todo app",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Elyro",
		Width:     1000,
		Height:    680,
		MinWidth:  640,
		MinHeight: 420,
		Mac: application.MacWindow{
			// Content runs under the title bar; the sidebar leaves room for the traffic lights.
			InvisibleTitleBarHeight: 52,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		URL: "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
