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
