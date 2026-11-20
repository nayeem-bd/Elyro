package main

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// Guards the pinned driver: fails if a modernc upgrade changes the bundled SQLite.
func TestSQLiteDriver(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var v string
	if err := db.QueryRow("select sqlite_version()").Scan(&v); err != nil {
		t.Fatal(err)
	}
	if v != "3.54.1" {
		t.Fatalf("sqlite_version() = %s, want 3.54.1", v)
	}
}
