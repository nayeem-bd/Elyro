package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	var applied *Settings
	s := &SettingsService{path: path, apply: func(v Settings) { applied = &v }}
	if got := s.Get(); got.HideDock {
		t.Errorf("missing file = %+v, want defaults", got)
	}
	if err := s.Set(Settings{HideDock: true}); err != nil {
		t.Fatal(err)
	}
	if !s.Get().HideDock || applied == nil || !applied.HideDock {
		t.Errorf("after Set: got %+v, applied %+v", s.Get(), applied)
	}
	os.WriteFile(path, []byte("{nope"), 0o600)
	if s.Get().HideDock {
		t.Error("corrupt file should give defaults")
	}
}
