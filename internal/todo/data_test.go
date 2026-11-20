package todo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportImport(t *testing.T) {
	ls, ts, inbox := services(t)
	data := &DataService{DB: ts.DB, BackupDir: t.TempDir()}
	work := must(ls.Create("Work"))
	must(ls.ShowInMenuBar(work.ID, true))
	a := must(ts.Create(inbox, "a"))
	must(ts.Update(a.ID, TaskInput{Title: "a", Notes: "n", DueDate: new("2026-12-01"), Priority: 2}))
	must(ts.SetDone(must(ts.Create(inbox, "done")).ID, true))
	must(ts.Create(work.ID, "w"))
	snapshot := func() string {
		var out []string
		for _, l := range must(ls.All()) {
			for _, t := range must(ts.ByList(l.ID)) {
				due := ""
				if t.DueDate != nil {
					due = *t.DueDate
				}
				out = append(out, fmt.Sprintf("%s|%v|%s|%s|%s|%d|%v", l.Name, l.InMenuBar, t.Title, t.Notes, due, t.Priority, t.CompletedAt != nil))
			}
		}
		return strings.Join(out, "\n")
	}
	want := snapshot()

	path := filepath.Join(t.TempDir(), "export.json")
	if err := data.Export(path); err != nil {
		t.Fatal(err)
	}
	must(ls.Create("Scratch")) // changes after the export are replaced by the import
	must(ts.Create(inbox, "later"))
	if err := data.Import(path); err != nil {
		t.Fatal(err)
	}
	if got := snapshot(); got != want {
		t.Errorf("after round trip:\n%s\nwant:\n%s", got, want)
	}
	if snaps, _ := filepath.Glob(filepath.Join(data.BackupDir, "before-import-*.db")); len(snaps) != 1 {
		t.Errorf("pre-import backups = %v, want 1", snaps)
	}

	file := func(lists string) string { return `{"app":"Elyro","version":1,"lists":[` + lists + `]}` }
	inInbox := func(task string) string { return file(`{"name":"Inbox","inbox":true,"tasks":[` + task + `]}`) }
	for name, body := range map[string]string{
		"not json":      `nope`,
		"other app":     `{"app":"Other","version":1,"lists":[]}`,
		"future":        `{"app":"Elyro","version":2,"lists":[]}`,
		"unknown field": file(`{"name":"Inbox","inbox":true,"tasks":[],"colour":"red"}`),
		"no inbox":      file(`{"name":"A","tasks":[]}`),
		"two inboxes":   file(`{"name":"A","inbox":true,"tasks":[]},{"name":"B","inbox":true,"tasks":[]}`),
		"blank title":   inInbox(`{"title":" "}`),
		"bad due":       inInbox(`{"title":"x","dueDate":"2026-02-30"}`),
		"bad priority":  inInbox(`{"title":"x","priority":7}`),
		"bad completed": inInbox(`{"title":"x","completedAt":"yesterday"}`),
	} {
		bad := filepath.Join(t.TempDir(), "bad.json")
		os.WriteFile(bad, []byte(body), 0o600)
		if err := data.Import(bad); err == nil {
			t.Errorf("%s: Import succeeded, want error", name)
		}
	}
	if got := snapshot(); got != want {
		t.Errorf("a rejected import changed data:\n%s", got)
	}

	// A minimal hand-written file: missing createdAt and notes are fine, text is trimmed.
	min := filepath.Join(t.TempDir(), "min.json")
	os.WriteFile(min, []byte(file(`{"name":" Inbox ","inbox":true,"tasks":[{"title":" hi "}]}`)), 0o600)
	if err := data.Import(min); err != nil {
		t.Fatal(err)
	}
	if got := snapshot(); got != "Inbox|false|hi|||0|false" {
		t.Errorf("minimal import = %q", got)
	}
}
