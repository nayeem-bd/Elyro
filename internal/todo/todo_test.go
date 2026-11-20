package todo

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"elyro/internal/store"
)

func services(t *testing.T) (*ListService, *TaskService, int64) {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "elyro.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ls, ts := &ListService{DB: db}, &TaskService{DB: db}
	all, err := ls.All()
	if err != nil || len(all) != 1 || !all[0].IsInbox {
		t.Fatalf("fresh DB lists = %+v, %v; want just the Inbox", all, err)
	}
	return ls, ts, all[0].ID
}

// must unwraps (value, error); a panic fails the test with a stack trace.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func TestLists(t *testing.T) {
	ls, _, inbox := services(t)

	work := must(ls.Create("  Work  "))
	home := must(ls.Create("Home"))
	if work.Name != "Work" || work.IsInbox {
		t.Errorf("Create trimmed = %+v", work)
	}
	all := must(ls.All())
	if len(all) != 3 || all[0].ID != inbox || all[1].ID != work.ID || all[2].ID != home.ID {
		t.Errorf("All order = %+v, want Inbox, Work, Home", all)
	}

	if !all[0].InMenuBar || all[1].InMenuBar {
		t.Errorf("menu bar defaults = %+v, want only the Inbox", all)
	}
	if got := must(ls.ShowInMenuBar(work.ID, true)); !got.InMenuBar {
		t.Errorf("ShowInMenuBar = %+v", got)
	}
	if _, err := ls.ShowInMenuBar(999, true); !errors.Is(err, ErrNotFound) {
		t.Errorf("ShowInMenuBar missing: err = %v, want ErrNotFound", err)
	}

	if got := must(ls.Rename(work.ID, "Job")); got.Name != "Job" {
		t.Errorf("Rename = %+v", got)
	}
	if _, err := ls.Rename(999, "x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Rename missing: err = %v, want ErrNotFound", err)
	}

	for _, bad := range []string{"", "   ", strings.Repeat("a", maxTextLen+1)} {
		if _, err := ls.Create(bad); err == nil {
			t.Errorf("Create(%.10q…) succeeded, want error", bad)
		}
		if _, err := ls.Rename(work.ID, bad); err == nil {
			t.Errorf("Rename(%.10q…) succeeded, want error", bad)
		}
	}

	if err := ls.Delete(work.ID); err != nil {
		t.Fatal(err)
	}
	if got := must(ls.All()); len(got) != 2 {
		t.Errorf("lists after Delete = %+v, want 2", got)
	}
	if err := ls.Delete(inbox); err == nil {
		t.Error("Delete(Inbox) succeeded, want error")
	}
	if err := ls.Delete(999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete missing: err = %v, want ErrNotFound", err)
	}
}

func TestTasks(t *testing.T) {
	_, ts, inbox := services(t)

	a := must(ts.Create(inbox, " Buy milk "))
	b := must(ts.Create(inbox, "Call mum"))
	c := must(ts.Create(inbox, "Pay rent"))
	if a.Title != "Buy milk" || a.ListID != inbox || a.CompletedAt != nil || a.DueDate != nil {
		t.Errorf("Create = %+v", a)
	}
	if _, err := ts.Create(inbox, "  "); err == nil {
		t.Error("Create blank succeeded, want error")
	}
	if _, err := ts.Create(999, "x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Create in missing list: err = %v, want ErrNotFound", err)
	}

	if got := must(ts.Update(b.ID, TaskInput{Title: "Call mom"})); got.Title != "Call mom" {
		t.Errorf("Update = %+v", got)
	}
	if _, err := ts.Update(b.ID, TaskInput{Title: ""}); err == nil {
		t.Error("Update blank succeeded, want error")
	}

	done := must(ts.SetDone(a.ID, true))
	if done.CompletedAt == nil {
		t.Fatal("SetDone(true) left completedAt nil")
	}
	const old = "2000-01-01T00:00:00Z" // an old stamp, so a refresh to now() would show
	must(ts.DB.Exec("UPDATE tasks SET completed_at = ? WHERE id = ?", old, a.ID))
	if again := must(ts.SetDone(a.ID, true)); *again.CompletedAt != old {
		t.Errorf("SetDone(true) on a done task changed completedAt to %s", *again.CompletedAt)
	}

	// Open tasks first in creation order, completed last.
	ids := func() (out []int64) {
		for _, x := range must(ts.ByList(inbox)) {
			out = append(out, x.ID)
		}
		return
	}
	if got := ids(); len(got) != 3 || got[0] != b.ID || got[1] != c.ID || got[2] != a.ID {
		t.Errorf("ByList order = %v, want [%d %d %d]", got, b.ID, c.ID, a.ID)
	}

	if open := must(ts.SetDone(a.ID, false)); open.CompletedAt != nil {
		t.Error("SetDone(false) kept completedAt")
	}

	if err := ts.Delete(c.ID); err != nil {
		t.Fatal(err)
	}
	if err := ts.Delete(c.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete twice: err = %v, want ErrNotFound", err)
	}
	for name, err := range map[string]error{
		"Update":  func() error { _, err := ts.Update(999, TaskInput{Title: "x"}); return err }(),
		"SetDone": func() error { _, err := ts.SetDone(999, true); return err }(),
	} {
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("%s missing: err = %v, want ErrNotFound", name, err)
		}
	}
}
