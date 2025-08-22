package student

import (
	"fmt"
	"strings"
)

// Columns lists all column names of the students table in a stable order.
// The order is used to build INSERT/SELECT/UPDATE statements.
var Columns = [...]string{
	"id",
	"first_name",
	"last_name",
	"grade",
	"created_at",
	"middle_name",
	"status",
	"home_address",
	"course_grades",
	"friends",
	"local",
	"exchange",
}

// ColumnsStr returns a comma-separated string of all columns in Columns.
func ColumnsStr() string {
	return strings.Join(Columns[:], ", ")
}

// placeholderFor returns a placeholder for a given column with proper casting.
// For JSONB columns we accept []byte and convert bytea -> text -> jsonb.
// NULL or empty becomes NULL.
func placeholderFor(col string) string {
	switch col {
	case "home_address", "course_grades", "local", "exchange":
		// text -> jsonb. NULL останется NULL без ошибки парсинга
		return fmt.Sprintf("CAST(:%s AS jsonb)", col)
	default:
		return ":" + col
	}
}

// NamedPlaceholders returns a parenthesized, comma-separated list of
// placeholders aligned with Columns, suitable for INSERT ... VALUES (...).
func NamedPlaceholders() string {
	parts := make([]string, 0, len(Columns))
	for _, col := range Columns {
		parts = append(parts, placeholderFor(col))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// UpdateSetStr returns the SET clause for UPDATE with named placeholders,
// skipping the primary key column "id".
func UpdateSetStr() string {
	sets := make([]string, 0, len(Columns)-1)
	for _, col := range Columns {
		if col == "id" {
			continue
		}
		sets = append(sets, fmt.Sprintf("%s = %s", col, placeholderFor(col)))
	}
	return strings.Join(sets, ", ")
}
