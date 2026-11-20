package todo

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func fixToday(t *testing.T) {
	t.Helper()
	clock = func() time.Time { return time.Date(2026, 11, 20, 9, 0, 0, 0, time.Local) }
	t.Cleanup(func() { clock = time.Now })
}

func titles(tasks []Task) string {
	var out []string
	for _, t := range tasks {
		out = append(out, t.Title)
	}
	return strings.Join(out, ",")
}

func TestUpdateFields(t *testing.T) {
	_, ts, inbox := services(t)
	a := must(ts.Create(inbox, "a"))

	got := must(ts.Update(a.ID, TaskInput{Title: "a", Notes: "  note  ", DueDate: new("2026-12-01"), Priority: 3}))
	if got.Notes != "note" || got.DueDate == nil || *got.DueDate != "2026-12-01" || got.Priority != 3 {
		t.Errorf("Update = %+v", got)
	}
	if got := must(ts.Update(a.ID, TaskInput{Title: "a", DueDate: new(" ")})); got.DueDate != nil {
		t.Errorf("blank due date should clear it, got %q", *got.DueDate)
	}

	for name, in := range map[string]TaskInput{
		"impossible day": {Title: "a", DueDate: new("2026-02-30")},
		"not a date":     {Title: "a", DueDate: new("tomorrow")},
		"unpadded":       {Title: "a", DueDate: new("2026-2-3")},
		"priority 4":     {Title: "a", Priority: 4},
		"priority -1":    {Title: "a", Priority: -1},
		"notes too long": {Title: "a", Notes: strings.Repeat("x", maxNotesLen+1)},
	} {
		if _, err := ts.Update(a.ID, in); err == nil {
			t.Errorf("%s: Update succeeded, want error", name)
		}
	}
}

func TestSmartViews(t *testing.T) {
	fixToday(t)
	ls, ts, inbox := services(t)
	work := must(ls.Create("Work"))

	add := func(list int64, title, due string, prio int) Task {
		task := must(ts.Create(list, title))
		var d *string
		if due != "" {
			d = &due
		}
		return must(ts.Update(task.ID, TaskInput{Title: title, DueDate: d, Priority: prio}))
	}
	add(inbox, "today-low", "2026-11-20", 1)
	add(work.ID, "overdue", "2026-11-19", 0)
	add(work.ID, "today-high", "2026-11-20", 3)
	add(inbox, "tomorrow", "2026-11-21", 0)
	add(inbox, "no-date", "", 0)
	done := add(inbox, "done-today", "2026-11-20", 0)
	must(ts.SetDone(done.ID, true))

	for name, c := range map[string]struct {
		fn   func() ([]Task, error)
		want string
	}{
		"Today":     {ts.Today, "overdue,today-high,today-low"},
		"Upcoming":  {ts.Upcoming, "tomorrow"},
		"Completed": {ts.Completed, "done-today"},
	} {
		if got := titles(must(c.fn())); got != c.want {
			t.Errorf("%s = %s, want %s", name, got, c.want)
		}
	}
}

func TestSearch(t *testing.T) {
	_, ts, inbox := services(t)
	a := must(ts.Create(inbox, "Buy MILK"))
	must(ts.Create(inbox, "100% done"))
	must(ts.Create(inbox, "snake_case"))
	must(ts.Create(inbox, "1000 things"))
	must(ts.Create(inbox, "snakeXcase"))
	must(ts.Update(a.ID, TaskInput{Title: "Buy MILK", Notes: "from the corner shop"}))

	for q, want := range map[string]string{
		"milk":   "Buy MILK", // case-insensitive
		"corner": "Buy MILK", // notes
		"%":      "100% done",
		"_":      "snake_case",
		"  ":     "",
		"zzz":    "",
	} {
		if got := titles(must(ts.Search(q))); got != want {
			t.Errorf("Search(%q) = %q, want %q", q, got, want)
		}
	}
}

func TestMoveTasks(t *testing.T) {
	ls, ts, inbox := services(t)
	work := must(ls.Create("Work"))
	a := must(ts.Create(inbox, "a"))
	b := must(ts.Create(inbox, "b"))
	c := must(ts.Create(inbox, "c"))
	w := must(ts.Create(work.ID, "w"))
	order := func(list int64) string { return titles(must(ts.ByList(list))) }

	must(ts.Move(c.ID, inbox, a.ID))
	if got := order(inbox); got != "c,a,b" {
		t.Errorf("move c before a: %s", got)
	}
	must(ts.Move(c.ID, inbox, 0))
	if got := order(inbox); got != "a,b,c" {
		t.Errorf("move c to end: %s", got)
	}
	must(ts.Move(a.ID, inbox, c.ID)) // downward, past b
	if got := order(inbox); got != "b,a,c" {
		t.Errorf("move a down before c: %s", got)
	}
	must(ts.Move(a.ID, inbox, b.ID)) // and back up
	if got := order(inbox); got != "a,b,c" {
		t.Errorf("move a back before b: %s", got)
	}
	moved := must(ts.Move(a.ID, work.ID, w.ID))
	if moved.ListID != work.ID || order(work.ID) != "a,w" || order(inbox) != "b,c" {
		t.Errorf("move a to Work before w: work=%s inbox=%s", order(work.ID), order(inbox))
	}

	if _, err := ts.Move(b.ID, inbox, w.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("beforeID in another list: err = %v, want ErrNotFound", err)
	}
	if _, err := ts.Move(b.ID, 999, 0); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing list: err = %v, want ErrNotFound", err)
	}
	if _, err := ts.Move(999, inbox, 0); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing task: err = %v, want ErrNotFound", err)
	}
}

func TestMoveLists(t *testing.T) {
	ls, _, inbox := services(t)
	a := must(ls.Create("A"))
	b := must(ls.Create("B"))
	names := func() string {
		var out []string
		for _, l := range must(ls.All()) {
			out = append(out, l.Name)
		}
		return strings.Join(out, ",")
	}

	must(ls.Move(b.ID, a.ID))
	if got := names(); got != "Inbox,B,A" {
		t.Errorf("move B before A: %s", got)
	}
	must(ls.Move(b.ID, 0))
	if got := names(); got != "Inbox,A,B" {
		t.Errorf("move B to end: %s", got)
	}
	if _, err := ls.Move(inbox, 0); err == nil {
		t.Error("moving the Inbox succeeded, want error")
	}
	if _, err := ls.Move(a.ID, inbox); !errors.Is(err, ErrNotFound) {
		t.Errorf("move before the Inbox: err = %v, want ErrNotFound", err)
	}
}
