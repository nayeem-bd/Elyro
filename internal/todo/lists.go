package todo

import (
	"database/sql"
	"errors"
)

type List struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	IsInbox bool   `json:"isInbox"`
}

type ListService struct{ DB *sql.DB }

const listCols = "id, name, is_inbox"

func scanList(row interface{ Scan(...any) error }) (List, error) {
	var l List
	err := row.Scan(&l.ID, &l.Name, &l.IsInbox)
	return l, err
}

// All returns the Inbox first, then user lists in their saved order.
func (s *ListService) All() ([]List, error) {
	rows, err := s.DB.Query("SELECT "+listCols+" FROM lists ORDER BY is_inbox DESC, position, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lists := []List{} // [] not null for the frontend
	for rows.Next() {
		l, err := scanList(rows)
		if err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

// Create adds a list at the end.
func (s *ListService) Create(name string) (List, error) {
	name, err := clean(name, "List name")
	if err != nil {
		return List{}, err
	}
	t := now()
	return scanList(s.DB.QueryRow(`INSERT INTO lists (name, position, created_at, updated_at)
		VALUES (?, (SELECT COALESCE(MAX(position), 0) + 1 FROM lists), ?, ?)
		RETURNING `+listCols, name, t, t))
}

func (s *ListService) Rename(id int64, name string) (List, error) {
	name, err := clean(name, "List name")
	if err != nil {
		return List{}, err
	}
	l, err := scanList(s.DB.QueryRow("UPDATE lists SET name = ?, updated_at = ? WHERE id = ? RETURNING "+listCols,
		name, now(), id))
	return l, notFound(err)
}

// Delete removes a list and (by cascade) its tasks. The Inbox cannot be deleted.
func (s *ListService) Delete(id int64) error {
	var inbox bool
	if err := s.DB.QueryRow("SELECT is_inbox FROM lists WHERE id = ?", id).Scan(&inbox); err != nil {
		return notFound(err)
	}
	if inbox {
		return errors.New("the Inbox cannot be deleted")
	}
	_, err := s.DB.Exec("DELETE FROM lists WHERE id = ?", id)
	return err
}
