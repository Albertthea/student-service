package student

import (
	"fmt"
	"strings"
)

// Columns holds the list of all column names in the students table.
var Columns = [...]string{"id", "first_name", "last_name", "grade", "created_at"}

// ColumnsStr returns a comma-separated string of all columns.
func ColumnsStr() string {
	return strings.Join(Columns[:], ", ")
}

// Placeholders generates a string of parameter placeholders for SQL queries.
func Placeholders(n int) string {
	params := make([]string, n)
	for i := 0; i < n; i++ {
		params[i] = fmt.Sprintf("$%d", i+1)
	}
	return "(" + strings.Join(params, ", ") + ")"
}

// NamedPlaceholders returns a string of named SQL placeholders for all columns.
func NamedPlaceholders() string {
	var placeholders []string
	for _, col := range Columns {
		placeholders = append(placeholders, ":"+col)
	}
	return "(" + strings.Join(placeholders, ", ") + ")"
}

// UpdateSetStr возвращает строку для SET в UPDATE запросе, например:
// "first_name = $1, last_name = $2, grade = $3, created_at = $4"
func UpdateSetStr() string {
	var sets []string
	for _, col := range Columns {
		if col == "id" {
			continue
		}
		sets = append(sets, fmt.Sprintf("%s = :%s", col, col))
	}
	return strings.Join(sets, ", ")
}
