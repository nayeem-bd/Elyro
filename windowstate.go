package main

import (
	"encoding/json"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	defaultWidth  = 1000
	defaultHeight = 680
	minWidth      = 640
	minHeight     = 420
)

// loadBounds reads the saved window frame. A missing, corrupt or too-small file gives ok=false.
func loadBounds(path string) (r application.Rect, ok bool) {
	b, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(b, &r) != nil || r.Width < minWidth || r.Height < minHeight {
		return application.Rect{}, false
	}
	return r, true
}

func saveBounds(path string, r application.Rect) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// settle takes the window's new frame, but keeps the previous size when it differs by only a pixel:
// Wails truncates the float frame to ints, so reading back a set size loses 1px, which would make the
// window shrink a little on every launch.
func settle(prev, now application.Rect) application.Rect {
	if d := now.Width - prev.Width; d >= -1 && d <= 1 {
		now.Width = prev.Width
	}
	if d := now.Height - prev.Height; d >= -1 && d <= 1 {
		now.Height = prev.Height
	}
	return now
}

// onScreen reports whether the window's title bar area lies on some screen, so a frame saved
// on a monitor that has since been unplugged isn't restored out of reach.
func onScreen(r application.Rect, screens []*application.Screen) bool {
	title := application.Point{X: r.X + r.Width/2, Y: r.Y + 20}
	for _, s := range screens {
		if s.Bounds.Contains(title) {
			return true
		}
	}
	return false
}
