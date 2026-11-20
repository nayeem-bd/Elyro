package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"elyro/internal/todo"
)

//go:embed build/tray/tray.png
var trayIcon []byte

// quickAddKey works from any app (Carbon hot key: no Accessibility permission needed).
const quickAddKey = "CmdOrCtrl+Option+N"

// menuBar adds the ✓ menu bar icon.
func menuBar(app *application.App, run func(cmd string), tasks *todo.TaskService) {
	menu := app.NewMenu()
	menu.Add("Quick Add Task…").SetAccelerator(quickAddKey).OnClick(send(run, "newTask"))
	today := menu.Add("Today").OnClick(send(run, "today"))
	menu.AddSeparator()
	login := menu.AddCheckbox("Launch at Login", false)
	syncLogin := func() { // show the real state, whatever the last click or System Settings did
		on, _ := app.Autostart.IsEnabled()
		login.SetChecked(on)
	}
	login.OnClick(func(*application.Context) {
		setLaunchAtLogin(app, login.Checked()) // the click has already toggled it
		syncLogin()
	})
	menu.AddSeparator()
	menu.AddRole(application.Quit) // "Quit Elyro"

	tray := app.SystemTray.New().SetTemplateIcon(trayIcon).SetMenu(menu)
	tray.SetTooltip("Elyro")
	// Refresh what the menu shows each time it opens, so nothing has to watch for changes.
	open := func() {
		today.SetLabel(todayLabel(tasks))
		syncLogin()
		tray.OpenMenu()
	}
	tray.OnClick(open).OnRightClick(open)

	if err := app.GlobalShortcut.Register(quickAddKey, func() { run("newTask") }); err != nil {
		log.Printf("global shortcut %s: %v", quickAddKey, err) // e.g. another app owns it; the menu item still works
	}
}

func todayLabel(tasks *todo.TaskService) string {
	due, err := tasks.Today()
	if err != nil || len(due) == 0 {
		return "Today"
	}
	return fmt.Sprintf("Today (%d)", len(due))
}

func setLaunchAtLogin(app *application.App, on bool) {
	var err error
	if on {
		err = app.Autostart.Enable()
	} else {
		err = app.Autostart.Disable()
	}
	if err != nil {
		app.Dialog.Error().SetTitle("Launch at Login").
			SetMessage(fmt.Sprintf("%v\n\nIf Elyro is turned off in System Settings › General › Login Items, turn it on there.", err)).
			Show()
	}
}
