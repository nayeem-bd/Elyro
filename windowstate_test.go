package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestBoundsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "window.json")
	if _, ok := loadBounds(path); ok {
		t.Error("missing file loaded as ok")
	}
	want := application.Rect{X: 10, Y: 20, Width: 900, Height: 700}
	if err := saveBounds(path, want); err != nil {
		t.Fatal(err)
	}
	if got, ok := loadBounds(path); !ok || got != want {
		t.Errorf("loadBounds = %+v, %v; want %+v", got, ok, want)
	}

	for name, body := range map[string]string{
		"corrupt":   "{nope",
		"too small": `{"X":0,"Y":0,"Width":100,"Height":100}`,
	} {
		os.WriteFile(path, []byte(body), 0o600)
		if _, ok := loadBounds(path); ok {
			t.Errorf("%s file loaded as ok", name)
		}
	}
}

func TestSettle(t *testing.T) {
	prev := application.Rect{X: 0, Y: 0, Width: 1000, Height: 680}
	if got := settle(prev, application.Rect{X: 5, Y: 6, Width: 999, Height: 679}); got != (application.Rect{X: 5, Y: 6, Width: 1000, Height: 680}) {
		t.Errorf("1px truncation not absorbed: %+v", got)
	}
	if got := settle(prev, application.Rect{Width: 900, Height: 700}); got.Width != 900 || got.Height != 700 {
		t.Errorf("real resize dropped: %+v", got)
	}
}

func TestOnScreen(t *testing.T) {
	screens := []*application.Screen{
		{Bounds: application.Rect{X: 0, Y: 0, Width: 1512, Height: 982}},
		{Bounds: application.Rect{X: 1512, Y: 0, Width: 2560, Height: 1440}},
	}
	for name, c := range map[string]struct {
		r    application.Rect
		want bool
	}{
		"main screen":       {application.Rect{X: 100, Y: 100, Width: 1000, Height: 680}, true},
		"second screen":     {application.Rect{X: 2000, Y: 200, Width: 1000, Height: 680}, true},
		"unplugged monitor": {application.Rect{X: -3000, Y: 100, Width: 1000, Height: 680}, false},
		"title bar above":   {application.Rect{X: 100, Y: -500, Width: 1000, Height: 680}, false},
	} {
		if got := onScreen(c.r, screens); got != c.want {
			t.Errorf("%s: onScreen = %v, want %v", name, got, c.want)
		}
	}
}
