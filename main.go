package main

import (
	"embed"
	"log"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"elyro/internal/store"
	"elyro/internal/todo"
)

//go:embed all:frontend/dist
var assets embed.FS

// The native menu sends its commands to the frontend as a "menu" event carrying the command name.
func init() { application.RegisterEvent[string]("menu") }

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

	statePath := filepath.Join(filepath.Dir(dbPath), "window.json")
	saved, restore := loadBounds(statePath)
	var (
		mu     sync.Mutex
		bounds = application.Rect{Width: defaultWidth, Height: defaultHeight} // latest frame, written once on quit
	)
	if restore {
		bounds = saved
	}

	app := application.New(application.Options{
		Name:        "Elyro",
		Description: "Offline todo app",
		Services: []application.Service{
			application.NewService(&todo.ListService{DB: db}),
			application.NewService(&todo.TaskService{DB: db}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		OnShutdown: func() {
			mu.Lock()
			defer mu.Unlock()
			if err := saveBounds(statePath, bounds); err != nil {
				log.Printf("save window state: %v", err)
			}
		},
	})
	app.Menu.Set(appMenu(app))

	opts := application.WebviewWindowOptions{
		Title:     "Elyro",
		Width:     bounds.Width,
		Height:    bounds.Height,
		MinWidth:  minWidth,
		MinHeight: minHeight,
		Mac: application.MacWindow{
			// Content runs under the title bar; the sidebar leaves room for the traffic lights.
			InvisibleTitleBarHeight: 52,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		URL: "/",
	}
	if app.Env.IsDarkMode() {
		opts.BackgroundColour = application.NewRGB(27, 29, 35) // ≈ the dark --color-white, so launch doesn't flash white
	}
	if restore {
		opts.InitialPosition, opts.X, opts.Y = application.WindowXY, saved.X, saved.Y
	}
	win := app.Window.NewWithOptions(opts)

	track := func(*application.WindowEvent) {
		mu.Lock()
		bounds = settle(bounds, win.Bounds())
		mu.Unlock()
	}
	win.RegisterHook(events.Common.WindowDidMove, track)
	win.RegisterHook(events.Common.WindowDidResize, track)
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		if restore && !onScreen(saved, app.Screen.GetAll()) {
			win.Center() // saved on a monitor that is gone
		}
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func appMenu(app *application.App) *application.Menu {
	send := func(cmd string) func(*application.Context) {
		return func(*application.Context) { app.Event.Emit("menu", cmd) }
	}
	menu := app.NewMenu()
	menu.AddRole(application.AppMenu)

	file := menu.AddSubmenu("File")
	file.Add("New Task").SetAccelerator("CmdOrCtrl+N").OnClick(send("newTask"))
	file.Add("New List").SetAccelerator("CmdOrCtrl+Shift+N").OnClick(send("newList"))
	file.AddSeparator()
	file.AddRole(application.CloseWindow) // the FileMenu role, minus its order: New items belong first

	menu.AddRole(application.EditMenu) // keeps copy/paste working in inputs
	edit := menu.FindByRole(application.EditMenu).GetSubmenu()
	edit.AddSeparator()
	edit.Add("Find").SetAccelerator("CmdOrCtrl+F").OnClick(send("find"))

	menu.AddRole(application.WindowMenu)
	return menu
}
