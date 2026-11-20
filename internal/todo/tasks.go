package todo

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

type Task struct {
	ID          int64   `json:"id"`
	ListID      int64   `json:"listId"`
	Title       string  `json:"title"`
	Notes       string  `json:"notes"`
	DueDate     *string `json:"dueDate"`     // 'YYYY-MM-DD' or null
	Priority    int     `json:"priority"`    // 0 none … 3 high
	CompletedAt *string `json:"completedAt"` // null = open
}

// TaskInput is what the user can edit.
type TaskInput struct {
	Title    string  `json:"title"`
	Notes    string  `json:"notes"`
	DueDate  *string `json:"dueDate"`
	Priority int     `json:"priority"`
}

type TaskService struct{ DB *sql.DB }

const taskCols = "id, list_id, title, notes, due_date, priority, completed_at"

func scanTask(row interface{ Scan(...any) error }) (Task, error) {
	var t Task
	err := row.Scan(&t.ID, &t.ListID, &t.Title, &t.Notes, &t.DueDate, &t.Priority, &t.CompletedAt)
	return t, err
}

// query runs SELECT taskCols FROM tasks WHERE where ORDER BY order; where/order are code constants.
func (s *TaskService) query(where, order string, args ...any) ([]Task, error) {
	rows, err := s.DB.Query("SELECT "+taskCols+" FROM tasks WHERE "+where+" ORDER BY "+order, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []Task{} // [] not null for the frontend
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

const byDue = "due_date, priority DESC, position, id"

// ByList returns a list's open tasks in order, then its completed ones.
func (s *TaskService) ByList(listID int64) ([]Task, error) {
	return s.query("list_id = ?", "completed_at IS NOT NULL, position, id", listID)
}

// Today returns open tasks due today or overdue, across all lists.
func (s *TaskService) Today() ([]Task, error) {
	return s.query("completed_at IS NULL AND due_date <= ?", byDue, today())
}

// Upcoming returns open tasks due after today.
func (s *TaskService) Upcoming() ([]Task, error) {
	return s.query("completed_at IS NULL AND due_date > ?", byDue, today())
}

// Completed returns done tasks, most recently completed first.
// shortcut: returns every completed task; add paging if this grows into the thousands.
func (s *TaskService) Completed() ([]Task, error) {
	return s.query("completed_at IS NOT NULL", "completed_at DESC, id DESC")
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// Search matches titles and notes (case-insensitive for ASCII), open tasks first.
func (s *TaskService) Search(q string) ([]Task, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []Task{}, nil
	}
	pattern := "%" + likeEscaper.Replace(q) + "%"
	return s.query(`(title LIKE ?1 ESCAPE '\' OR notes LIKE ?1 ESCAPE '\')`,
		"completed_at IS NOT NULL, updated_at DESC, id DESC", pattern)
}

// Create adds a task at the end of a list.
func (s *TaskService) Create(listID int64, title string) (Task, error) {
	title, err := clean(title, "Task title")
	if err != nil {
		return Task{}, err
	}
	if err := requireList(s.DB, listID); err != nil {
		return Task{}, err
	}
	t := now()
	return scanTask(s.DB.QueryRow(`INSERT INTO tasks (list_id, title, position, created_at, updated_at)
		VALUES (?1, ?2, (SELECT COALESCE(MAX(position), 0) + 1 FROM tasks WHERE list_id = ?1), ?3, ?3)
		RETURNING `+taskCols, listID, title, t))
}

// clean validates and normalises editable fields (shared by Update and Import).
func (in TaskInput) clean() (TaskInput, error) {
	var err error
	if in.Title, err = clean(in.Title, "Task title"); err != nil {
		return in, err
	}
	in.Notes = strings.TrimSpace(in.Notes)
	if utf8.RuneCountInString(in.Notes) > maxNotesLen {
		return in, fmt.Errorf("notes are too long (max %d characters)", maxNotesLen)
	}
	if in.DueDate, err = cleanDue(in.DueDate); err != nil {
		return in, err
	}
	if in.Priority < 0 || in.Priority > 3 {
		return in, errors.New("priority must be 0 (none) to 3 (high)")
	}
	return in, nil
}

// Update replaces a task's editable fields.
func (s *TaskService) Update(id int64, in TaskInput) (Task, error) {
	in, err := in.clean()
	if err != nil {
		return Task{}, err
	}
	t, err := scanTask(s.DB.QueryRow(`UPDATE tasks SET title = ?, notes = ?, due_date = ?, priority = ?, updated_at = ?
		WHERE id = ? RETURNING `+taskCols, in.Title, in.Notes, in.DueDate, in.Priority, now(), id))
	return t, notFound(err)
}

// SetDone marks a task done or open. Marking a done task done again keeps its original completion time.
func (s *TaskService) SetDone(id int64, done bool) (Task, error) {
	t := now()
	task, err := scanTask(s.DB.QueryRow(`UPDATE tasks
		SET completed_at = CASE WHEN ?1 THEN COALESCE(completed_at, ?2) END, updated_at = ?2
		WHERE id = ?3 RETURNING `+taskCols, done, t, id))
	return task, notFound(err)
}

// Move puts a task into listID just before beforeID (0 = at the end of the list).
func (s *TaskService) Move(id, listID, beforeID int64) (Task, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return Task{}, err
	}
	defer tx.Rollback()
	if err := requireList(tx, listID); err != nil {
		return Task{}, err
	}
	pos, err := placeBefore(tx, "tasks", "list_id = ?", []any{listID}, id, beforeID)
	if err != nil {
		return Task{}, err
	}
	task, err := scanTask(tx.QueryRow("UPDATE tasks SET list_id = ?, position = ?, updated_at = ? WHERE id = ? RETURNING "+taskCols,
		listID, pos, now(), id))
	if err != nil {
		return Task{}, notFound(err)
	}
	return task, tx.Commit()
}

func (s *TaskService) Delete(id int64) error {
	var gone int64
	return notFound(s.DB.QueryRow("DELETE FROM tasks WHERE id = ? RETURNING id", id).Scan(&gone))
}
