package store

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func openTemp(t *testing.T) (*sql.DB, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sub dir", "elyro.db") // space + missing dir on purpose
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db, path
}

func queryString(t *testing.T, db *sql.DB, q string) string {
	t.Helper()
	var s string
	if err := db.QueryRow(q).Scan(&s); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return s
}

func TestOpenFreshDB(t *testing.T) {
	db, _ := openTemp(t)
	for q, want := range map[string]string{
		"select sqlite_version()":                      "3.54.1", // guards the pinned driver
		"PRAGMA user_version":                          "1",
		"PRAGMA foreign_keys":                          "1",
		"PRAGMA journal_mode":                          "wal",
		"select count(*) from lists where is_inbox = 1": "1",
		"select name from lists where is_inbox = 1":     "Inbox",
	} {
		if got := queryString(t, db, q); got != want {
			t.Errorf("%s = %q, want %q", q, got, want)
		}
	}
}

func TestReopenIsIdempotent(t *testing.T) {
	db, path := openTemp(t)
	db.Close()
	db2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	if got := queryString(t, db2, "select count(*) from lists"); got != "1" {
		t.Errorf("lists after reopen = %s, want 1 (Inbox seeded once)", got)
	}
	if got := queryString(t, db2, "PRAGMA user_version"); got != "1" {
		t.Errorf("user_version after reopen = %s, want 1", got)
	}
}

const now = "'2026-11-20T00:00:00Z'"

func TestConstraintsRejectBadData(t *testing.T) {
	db, _ := openTemp(t)
	for name, stmt := range map[string]string{
		"blank list name":  "insert into lists (name, position, created_at, updated_at) values ('  ', 1, " + now + ", " + now + ")",
		"second inbox":     "insert into lists (name, is_inbox, position, created_at, updated_at) values ('X', 1, 1, " + now + ", " + now + ")",
		"blank task title": "insert into tasks (list_id, title, position, created_at, updated_at) values (1, ' ', 1, " + now + ", " + now + ")",
		"priority 4":       "insert into tasks (list_id, title, priority, position, created_at, updated_at) values (1, 'a', 4, 1, " + now + ", " + now + ")",
		"missing list":     "insert into tasks (list_id, title, position, created_at, updated_at) values (999, 'a', 1, " + now + ", " + now + ")",
	} {
		if _, err := db.Exec(stmt); err == nil {
			t.Errorf("%s: insert succeeded, want constraint error", name)
		}
	}
}

func TestDeletingListDeletesItsTasks(t *testing.T) {
	db, _ := openTemp(t)
	mustExec(t, db, "insert into lists (id, name, position, created_at, updated_at) values (2, 'Work', 1, "+now+", "+now+")")
	mustExec(t, db, "insert into tasks (list_id, title, position, created_at, updated_at) values (2, 'a', 1, "+now+", "+now+")")
	mustExec(t, db, "delete from lists where id = 2")
	if got := queryString(t, db, "select count(*) from tasks"); got != "0" {
		t.Errorf("tasks after list delete = %s, want 0", got)
	}
}

func mustExec(t *testing.T, db *sql.DB, q string) {
	t.Helper()
	if _, err := db.Exec(q); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
}
