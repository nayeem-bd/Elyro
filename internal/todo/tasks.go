package todo

import "database/sql"

type Task struct {
	ID          int64   `json:"id"`
	ListID      int64   `json:"listId"`
	Title       string  `json:"title"`
	Notes       string  `json:"notes"`
	DueDate     *string `json:"dueDate"`     // 'YYYY-MM-DD' or null
	Priority    int     `json:"priority"`    // 0 none … 3 high
	CompletedAt *string `json:"completedAt"` // null = open
}

// TaskInput is what the user can edit. Phase 3 adds notes, due date and priority.
type TaskInput struct {
	Title string `json:"title"`
}

type TaskService struct{ DB *sql.DB }

const taskCols = "id, list_id, title, notes, due_date, priority, completed_at"

func scanTask(row interface{ Scan(...any) error }) (Task, error) {
	var t Task
	err := row.Scan(&t.ID, &t.ListID, &t.Title, &t.Notes, &t.DueDate, &t.Priority, &t.CompletedAt)
	return t, err
}

// ByList returns a list's open tasks in order, then its completed ones.
func (s *TaskService) ByList(listID int64) ([]Task, error) {
	rows, err := s.DB.Query("SELECT "+taskCols+" FROM tasks WHERE list_id = ? ORDER BY completed_at IS NOT NULL, position, id", listID)
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

// Create adds a task at the end of a list.
func (s *TaskService) Create(listID int64, title string) (Task, error) {
	title, err := clean(title, "Task title")
	if err != nil {
		return Task{}, err
	}
	var exists bool
	if err := s.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM lists WHERE id = ?)", listID).Scan(&exists); err != nil {
		return Task{}, err
	}
	if !exists {
		return Task{}, ErrNotFound
	}
	t := now()
	return scanTask(s.DB.QueryRow(`INSERT INTO tasks (list_id, title, position, created_at, updated_at)
		VALUES (?1, ?2, (SELECT COALESCE(MAX(position), 0) + 1 FROM tasks WHERE list_id = ?1), ?3, ?3)
		RETURNING `+taskCols, listID, title, t))
}

func (s *TaskService) Update(id int64, in TaskInput) (Task, error) {
	title, err := clean(in.Title, "Task title")
	if err != nil {
		return Task{}, err
	}
	t, err := scanTask(s.DB.QueryRow("UPDATE tasks SET title = ?, updated_at = ? WHERE id = ? RETURNING "+taskCols,
		title, now(), id))
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

func (s *TaskService) Delete(id int64) error {
	var gone int64
	return notFound(s.DB.QueryRow("DELETE FROM tasks WHERE id = ? RETURNING id", id).Scan(&gone))
}
