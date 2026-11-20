package main

import (
	"encoding/json"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Settings are app preferences (not lists or tasks), kept as JSON next to the database.
type Settings struct {
	HideDock bool `json:"hideDock"` // run as a menu bar app with no Dock icon
}

// SettingsService reads and saves Settings; apply puts a saved change into effect right away.
type SettingsService struct {
	path  string
	apply func(Settings)
}

// Get returns the saved settings, or the defaults if the file is missing or unreadable.
func (s *SettingsService) Get() Settings {
	var v Settings
	if b, err := os.ReadFile(s.path); err == nil {
		_ = json.Unmarshal(b, &v)
	}
	return v
}

func (s *SettingsService) Set(v Settings) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.path, b, 0o600); err != nil {
		return err
	}
	s.apply(v)
	return nil
}

// dockPolicy is the launch-time form of HideDock (dock.HideAppIcon/ShowAppIcon switch it while running).
func dockPolicy(hide bool) application.ActivationPolicy {
	if hide {
		return application.ActivationPolicyAccessory
	}
	return application.ActivationPolicyRegular
}
