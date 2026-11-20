package main

import (
	"embed"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"elyro/internal/store"
	"elyro/internal/todo"
)

//go:embed all:frontend/dist
var assets embed.FS

// The native menus send commands to the frontend as a "menu" event: a command name, or a view path like "/today".
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

	dataDir := filepath.Dir(dbPath)
	backupDir := filepath.Join(dataDir, "Backups")
	// shortcut: one backup per launch-day; an app left open past midnight backs up on its next launch.
	if err := store.Backup(db, backupDir, time.Now().Format("2006-01-02"), 7); err != nil {
		log.Printf("daily backup: %v", err)
	}

	statePath := filepath.Join(dataDir, "window.json")
	saved, restore := loadBounds(statePath)
	var (
		mu     sync.Mutex
		bounds = application.Rect{Width: defaultWidth, Height: defaultHeight} // latest frame, written once on quit
	)
	if restore {
		bounds = saved
	}

	lists := &todo.ListService{DB: db}
	tasks := &todo.TaskService{DB: db}
	app := application.New(application.Options{
		Name:        "Elyro",
		Description: "Offline todo app",
		Services: []application.Service{
			application.NewService(lists),
			application.NewService(tasks),
			application.NewService(&todo.DataService{DB: db, BackupDir: backupDir}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		OnShutdown: func() {
			mu.Lock()
			defer mu.Unlock()
			if err := saveBounds(statePath, bounds); err != nil {
				log.Printf("save window state: %v", err)
			}
		},
	})
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
	// Started by Launch at Login: stay in the background. (macOS can't tell a login launch from a
	// manual one, so with it on, every cold start is hidden; the Dock or menu bar shows the window.)
	opts.Hidden, _ = app.Autostart.IsEnabled()
	win := app.Window.NewWithOptions(opts)

	// Closing the window only hides it; Elyro keeps running (⌘Q quits). Wails shows it again on a Dock click.
	win.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		win.Hide()
		e.Cancel()
	})
	run := func(cmd string) { // commands are handled in frontend/src/shortcuts.ts
		win.Show().Focus()
		app.Event.Emit("menu", cmd)
	}
	app.Menu.Set(appMenu(app, run))
	menuBar(app, run, lists, tasks)

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

// send makes a menu click handler that runs a frontend command.
func send(run func(cmd string), cmd string) func(*application.Context) {
	return func(*application.Context) { run(cmd) }
}

func appMenu(app *application.App, run func(cmd string)) *application.Menu {
	menu := app.NewMenu()
	menu.AddRole(application.AppMenu)

	file := menu.AddSubmenu("File")
	file.Add("New Task").SetAccelerator("CmdOrCtrl+N").OnClick(send(run, "newTask"))
	file.Add("New List").SetAccelerator("CmdOrCtrl+Shift+N").OnClick(send(run, "newList"))
	file.AddSeparator()
	file.Add("Export…").SetAccelerator("CmdOrCtrl+Shift+E").OnClick(send(run, "export"))
	file.Add("Import…").OnClick(send(run, "import"))
	file.AddSeparator()
	file.AddRole(application.CloseWindow) // the FileMenu role, minus its order: New items belong first

	menu.AddRole(application.EditMenu) // keeps copy/paste working in inputs
	edit := menu.FindByRole(application.EditMenu).GetSubmenu()
	edit.AddSeparator()
	edit.Add("Find").SetAccelerator("CmdOrCtrl+F").OnClick(send(run, "find"))

	menu.AddRole(application.WindowMenu)
	return menu
}
