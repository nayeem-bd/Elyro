package store

import (
	"database/sql"
	"os"
	"path/filepath"
)

// Snapshot writes a consistent copy of the database to path, replacing any file already there.
func Snapshot(db *sql.DB, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	_, err := db.Exec("VACUUM INTO ?", path)
	return err
}

// Backup snapshots the database to dir/elyro-<day>.db unless that day's copy exists,
// then deletes all but the newest keep daily copies. day is 'YYYY-MM-DD', so names sort by date.
func Backup(db *sql.DB, dir, day string, keep int) error {
	path := filepath.Join(dir, "elyro-"+day+".db")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := Snapshot(db, path); err != nil {
			return err
		}
	}
	return Prune(filepath.Join(dir, "elyro-????-??-??.db"), keep)
}

// Prune deletes all but the last keep files matching pattern, in name order
// (callers put a sortable timestamp in the name, so that is oldest first).
func Prune(pattern string, keep int) error {
	old, err := filepath.Glob(pattern) // sorted
	if err != nil || len(old) <= keep {
		return err
	}
	for _, p := range old[:len(old)-keep] {
		if err := os.Remove(p); err != nil {
			return err
		}
	}
	return nil
}
