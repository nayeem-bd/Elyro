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
func menuBar(app *application.App, run func(cmd string), lists *todo.ListService, tasks *todo.TaskService) {
	menu := app.NewMenu()
	tray := app.SystemTray.New().SetTemplateIcon(trayIcon).SetMenu(menu)
	tray.SetTooltip("Elyro")
	// Rebuilt each time it opens, so counts, list names and the lists picked in Settings are never stale.
	open := func() {
		menu.Clear()
		fillMenuBar(app, menu, run, lists, tasks)
		tray.SetMenu(menu) // re-renders the same native menu
		tray.OpenMenu()
	}
	tray.OnClick(open).OnRightClick(open)

	if err := app.GlobalShortcut.Register(quickAddKey, func() { run("newTask") }); err != nil {
		log.Printf("global shortcut %s: %v", quickAddKey, err) // e.g. another app owns it; the menu item still works
	}
}

// fillMenuBar: Quick Add, Today, the lists shown in the menu bar (Settings), then app items.
// A list or path command ("/today", "/list/2") opens that view in the window.
func fillMenuBar(app *application.App, menu *application.Menu, run func(cmd string), lists *todo.ListService, tasks *todo.TaskService) {
	menu.Add("Quick Add Task…").SetAccelerator(quickAddKey).OnClick(send(run, "newTask"))
	counts, err := tasks.Counts()
	if err != nil {
		log.Printf("menu bar counts: %v", err)
	}
	menu.Add(counted("Today", counts.Today)).OnClick(send(run, "/today"))

	all, err := lists.All()
	if err != nil {
		log.Printf("menu bar lists: %v", err)
	}
	var pinned []todo.List
	for _, l := range all {
		if l.InMenuBar {
			pinned = append(pinned, l)
		}
	}
	if len(pinned) > 0 {
		menu.AddSeparator()
		for _, l := range pinned {
			menu.Add(counted(l.Name, counts.Lists[l.ID])).OnClick(send(run, fmt.Sprintf("/list/%d", l.ID)))
		}
	}

	menu.AddSeparator()
	on, _ := app.Autostart.IsEnabled()
	login := menu.AddCheckbox("Launch at Login", on)
	login.OnClick(func(*application.Context) { setLaunchAtLogin(app, login.Checked()) }) // the click has already toggled it
	menu.Add("Settings…").OnClick(send(run, "/settings"))
	menu.AddSeparator()
	menu.AddRole(application.Quit) // "Quit Elyro"
}

// counted is "Name (n)", or just "Name" when there is nothing open.
func counted(name string, n int) string {
	if n == 0 {
		return name
	}
	return fmt.Sprintf("%s (%d)", name, n)
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
