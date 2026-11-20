package todo

import "database/sql"

// placeBefore returns the position that puts row id just before beforeID (0 = at the end)
// among the rows of table matching scope, shifting the rows from beforeID onward down by one.
// table and scope are code constants, never user input.
func placeBefore(tx *sql.Tx, table, scope string, scopeArgs []any, id, beforeID int64) (float64, error) {
	from := " FROM " + table + " WHERE " + scope + " AND id != ?"
	args := append(append([]any{}, scopeArgs...), id)

	var pos sql.NullFloat64
	if beforeID == 0 {
		err := tx.QueryRow("SELECT MAX(position)"+from, args...).Scan(&pos)
		return pos.Float64 + 1, err // empty scope: NULL → 0 + 1
	}
	if err := tx.QueryRow("SELECT position"+from+" AND id = ?", append(args, beforeID)...).Scan(&pos); err != nil {
		return 0, notFound(err) // beforeID missing or outside the scope
	}
	_, err := tx.Exec("UPDATE "+table+" SET position = position + 1 WHERE "+scope+" AND id != ? AND position >= ?",
		append(args, pos.Float64)...)
	return pos.Float64, err
}
