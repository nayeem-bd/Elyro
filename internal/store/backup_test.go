package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackup(t *testing.T) {
	db, _ := openTemp(t)
	dir := filepath.Join(t.TempDir(), "Backups")
	days := []string{"2026-11-01", "2026-11-02", "2026-11-03", "2026-11-04"}
	for _, d := range days {
		if err := Backup(db, dir, d, 2); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(dir, "notes.txt"), nil, 0o600) // not ours: must survive pruning
	mustExec(t, db, "INSERT INTO lists (name, position, created_at, updated_at) VALUES ('Later', 9, 'x', 'x')")
	if err := Backup(db, dir, "2026-11-04", 2); err != nil { // same day again: no new copy
		t.Fatal(err)
	}

	var names []string
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if got := strings.Join(names, ","); got != "elyro-2026-11-03.db,elyro-2026-11-04.db,notes.txt" {
		t.Errorf("backups = %s", got)
	}

	copy, err := Open(filepath.Join(dir, "elyro-2026-11-04.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer copy.Close()
	if got := queryString(t, copy, "select group_concat(name) from lists"); got != "Inbox" {
		t.Errorf("backup lists = %s, want the day's first snapshot (Inbox only)", got)
	}
}

func TestPrune(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"before-import-2026-11-01-090000.db", "before-import-2026-11-02-090000.db", "before-import-2026-11-02-100000.db", "keep.txt"} {
		os.WriteFile(filepath.Join(dir, n), nil, 0o600)
	}
	if err := Prune(filepath.Join(dir, "before-import-*.db"), 2); err != nil {
		t.Fatal(err)
	}
	var names []string
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if got := strings.Join(names, ","); got != "before-import-2026-11-02-090000.db,before-import-2026-11-02-100000.db,keep.txt" {
		t.Errorf("after prune = %s", got)
	}
}
