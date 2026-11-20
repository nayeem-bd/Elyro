// Package todo holds the list and task services bound to the frontend.
// All data rules live here; the frontend only displays and edits.
package todo

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrNotFound = errors.New("not found")

const (
	maxTextLen  = 500
	maxNotesLen = 10000
	dateLayout  = "2006-01-02"
)

var clock = time.Now // swapped in tests

// now is a UTC timestamp; today is the user's local calendar day (due dates are days, not instants).
func now() string   { return clock().UTC().Format(time.RFC3339) }
func today() string { return clock().Format(dateLayout) }

// cleanDue accepts nil/blank (no due date) or a real 'YYYY-MM-DD' day.
func cleanDue(d *string) (*string, error) {
	if d == nil || strings.TrimSpace(*d) == "" {
		return nil, nil
	}
	if _, err := time.Parse(dateLayout, *d); err != nil {
		return nil, fmt.Errorf("due date %q must be a valid YYYY-MM-DD day", *d)
	}
	return d, nil
}

type querier interface {
	QueryRow(query string, args ...any) *sql.Row
}

// refuseInbox fails if list id is missing (ErrNotFound) or is the Inbox, which can't be deleted or moved.
func refuseInbox(q querier, id int64, verb string) error {
	var inbox bool
	if err := q.QueryRow("SELECT is_inbox FROM lists WHERE id = ?", id).Scan(&inbox); err != nil {
		return notFound(err)
	}
	if inbox {
		return fmt.Errorf("the Inbox cannot be %s", verb)
	}
	return nil
}

func requireList(q querier, id int64) error {
	var exists bool
	if err := q.QueryRow("SELECT EXISTS (SELECT 1 FROM lists WHERE id = ?)", id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

// clean trims user text and rejects blank or oversized values (frontend input is untrusted).
func clean(s, what string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("%s cannot be empty", what)
	}
	if utf8.RuneCountInString(s) > maxTextLen {
		return "", fmt.Errorf("%s is too long (max %d characters)", what, maxTextLen)
	}
	return s, nil
}

func notFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
