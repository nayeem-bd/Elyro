package todo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"elyro/internal/store"
)

const exportVersion = 1

// The JSON export: every list (Inbox first) with all of its tasks, in display order.
type exportFile struct {
	App        string       `json:"app"`
	Version    int          `json:"version"`
	ExportedAt string       `json:"exportedAt"`
	Lists      []exportList `json:"lists"`
}

type exportList struct {
	Name  string       `json:"name"`
	Inbox bool         `json:"inbox,omitempty"`
	Tasks []exportTask `json:"tasks"`
}

type exportTask struct {
	TaskInput
	CompletedAt *string `json:"completedAt"`
	CreatedAt   string  `json:"createdAt"`
}

// DataService exports and imports everything as one JSON file.
type DataService struct {
	DB        *sql.DB
	BackupDir string // Import snapshots the current data here first
}

// Export writes all lists and tasks to path.
func (s *DataService) Export(path string) error {
	f := exportFile{App: "Elyro", Version: exportVersion, ExportedAt: now(), Lists: []exportList{}}
	lists, err := (&ListService{DB: s.DB}).All()
	if err != nil {
		return err
	}
	for _, l := range lists {
		tasks, err := s.exportTasks(l.ID)
		if err != nil {
			return err
		}
		f.Lists = append(f.Lists, exportList{Name: l.Name, Inbox: l.IsInbox, Tasks: tasks})
	}
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

// exportTasks returns a list's tasks in stored order (open and done mixed, as positioned).
func (s *DataService) exportTasks(listID int64) ([]exportTask, error) {
	rows, err := s.DB.Query(`SELECT title, notes, due_date, priority, completed_at, created_at
		FROM tasks WHERE list_id = ? ORDER BY position, id`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []exportTask{}
	for rows.Next() {
		var t exportTask
		if err := rows.Scan(&t.Title, &t.Notes, &t.DueDate, &t.Priority, &t.CompletedAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// Import replaces all lists and tasks with the contents of an export file. The file is fully
// validated first, and the current data is snapshotted to BackupDir/before-import-<time>.db.
func (s *DataService) Import(path string) error {
	f, err := readExport(path)
	if err != nil {
		return err
	}
	// One copy per import, so importing twice still leaves the data from before the first.
	snap := filepath.Join(s.BackupDir, "before-import-"+clock().Format("2006-01-02-150405")+".db")
	if err := store.Snapshot(s.DB, snap); err != nil {
		return fmt.Errorf("back up current data: %w", err)
	}
	if err := store.Prune(filepath.Join(s.BackupDir, "before-import-*.db"), keepImportBackups); err != nil {
		return fmt.Errorf("prune import backups: %w", err)
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stamp := now()
	if _, err := tx.Exec("DELETE FROM lists"); err != nil { // tasks go with them (ON DELETE CASCADE)
		return err
	}
	for i, l := range f.Lists {
		res, err := tx.Exec("INSERT INTO lists (name, is_inbox, position, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			l.Name, l.Inbox, i+1, stamp, stamp)
		if err != nil {
			return err
		}
		listID, _ := res.LastInsertId()
		for j, t := range l.Tasks {
			if _, err := tx.Exec(`INSERT INTO tasks (list_id, title, notes, due_date, priority, completed_at, position, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				listID, t.Title, t.Notes, t.DueDate, t.Priority, t.CompletedAt, j+1, t.CreatedAt, stamp); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

const (
	maxImportBytes    = 50 << 20
	keepImportBackups = 5
)

// readExport parses and validates an export file, normalising it the same way the services do.
func readExport(path string) (exportFile, error) {
	var f exportFile
	r, err := os.Open(path)
	if err != nil {
		return f, err
	}
	defer r.Close()
	dec := json.NewDecoder(io.LimitReader(r, maxImportBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return f, fmt.Errorf("not an Elyro export: %w", err)
	}
	if f.App != "Elyro" || f.Version != exportVersion {
		return f, errors.New("not an Elyro export (version 1)")
	}

	inboxes := 0
	for i := range f.Lists {
		l := &f.Lists[i]
		if l.Inbox {
			inboxes++
		}
		if l.Name, err = clean(l.Name, "List name"); err != nil {
			return f, fmt.Errorf("list %d: %w", i+1, err)
		}
		for j := range l.Tasks {
			if err := cleanImported(&l.Tasks[j]); err != nil {
				return f, fmt.Errorf("list %q, task %d: %w", l.Name, j+1, err)
			}
		}
	}
	if inboxes != 1 {
		return f, fmt.Errorf("the file must have exactly one Inbox, found %d", inboxes)
	}
	return f, nil
}

func cleanImported(t *exportTask) (err error) {
	if t.TaskInput, err = t.TaskInput.clean(); err != nil {
		return err
	}
	if t.CompletedAt != nil {
		if _, err := time.Parse(time.RFC3339, *t.CompletedAt); err != nil {
			return fmt.Errorf("completedAt %q is not an RFC 3339 time", *t.CompletedAt)
		}
	}
	if _, err := time.Parse(time.RFC3339, t.CreatedAt); err != nil {
		t.CreatedAt = now() // optional in hand-written files
	}
	return nil
}
