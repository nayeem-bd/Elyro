package todo

import "database/sql"

type List struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	IsInbox   bool   `json:"isInbox"`
	InMenuBar bool   `json:"inMenuBar"`
}

type ListService struct{ DB *sql.DB }

const listCols = "id, name, is_inbox, in_menu_bar"

func scanList(row interface{ Scan(...any) error }) (List, error) {
	var l List
	err := row.Scan(&l.ID, &l.Name, &l.IsInbox, &l.InMenuBar)
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

// ShowInMenuBar sets whether the list appears in the menu bar menu.
func (s *ListService) ShowInMenuBar(id int64, on bool) (List, error) {
	l, err := scanList(s.DB.QueryRow("UPDATE lists SET in_menu_bar = ?, updated_at = ? WHERE id = ? RETURNING "+listCols,
		on, now(), id))
	return l, notFound(err)
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
	if err := refuseInbox(s.DB, id, "deleted"); err != nil {
		return err
	}
	_, err := s.DB.Exec("DELETE FROM lists WHERE id = ?", id)
	return err
}

// Move puts a list just before beforeID (0 = at the end). The Inbox always stays first.
func (s *ListService) Move(id, beforeID int64) (List, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return List{}, err
	}
	defer tx.Rollback()
	if err := refuseInbox(tx, id, "moved"); err != nil {
		return List{}, err
	}
	pos, err := placeBefore(tx, "lists", "is_inbox = 0", nil, id, beforeID)
	if err != nil {
		return List{}, err
	}
	l, err := scanList(tx.QueryRow("UPDATE lists SET position = ?, updated_at = ? WHERE id = ? RETURNING "+listCols, pos, now(), id))
	if err != nil {
		return List{}, err
	}
	return l, tx.Commit()
}
