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

const maxTextLen = 500

func now() string { return time.Now().UTC().Format(time.RFC3339) }

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
