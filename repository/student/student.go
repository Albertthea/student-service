// Package student provides the data structures and repository methods
// for managing student records.
package student

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"example.com/student-service/internal/txmanager"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const tableName = "students"

var (
	// ErrNotFound is returned when a student record is not found in the database.
	ErrNotFound = errors.New("student not found")
	// ErrAlreadyExists is returned when trying to create a student that already exists.
	ErrAlreadyExists = errors.New("student already exists")
)

// Student represents a student record in the database.
type Student struct {
	ID           string         `db:"id" json:"id"`
	FirstName    string         `db:"first_name" json:"first_name"`
	LastName     string         `db:"last_name" json:"last_name"`
	Grade        int32          `db:"grade" json:"grade"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	MiddleName   sql.NullString `db:"middle_name" json:"middle_name"`
	Status       sql.NullString `db:"status" json:"status"`
	HomeAddress  []byte         `db:"home_address" json:"home_address"`
	CourseGrades []byte         `db:"course_grades" json:"course_grades"`
	Friends      pq.StringArray `db:"friends" json:"friends"`
	Local        []byte         `db:"local" json:"local"`
	Exchange     []byte         `db:"exchange" json:"exchange"`
}

// Repository manages operations with students in the DB.
type Repository struct {
	db *sqlx.DB
}

// DB returns the underlying sqlx.DB instance.
func (r *Repository) DB() *sqlx.DB { return r.db }

// NewRepository creates a new Repository with the given DB connection.
func NewRepository(db *sqlx.DB) *Repository { return &Repository{db: db} }

// Create inserts a new student record into the database.
func (r *Repository) Create(ctx context.Context, s Student) (string, error) {
	if s.ID == "" {
		return "", fmt.Errorf("create student: ID must be specified")
	}
	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		return "", fmt.Errorf("create student: tx required: %w", err)
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s`, tableName, ColumnsStr(), NamedPlaceholders())

	// Key fix: bind via map with JSON as string (or nil) so CAST(:col AS jsonb) works.
	if _, err = tx.NamedExecContext(ctx, query, toNamedArgs(s)); err != nil {
		return "", fmt.Errorf("create student: insert: %w", err)
	}
	return s.ID, nil
}

// GetByID retrieves a student by their ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*Student, error) {
	var s Student
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE id = $1`, ColumnsStr(), tableName)

	if tx, err := txmanager.GetTx(ctx); err == nil {
		if err := tx.GetContext(ctx, &s, query, id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("get student by id (tx): %w", err)
		}
	} else {
		err = r.db.GetContext(ctx, &s, query, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("get student by id: %w", err)
		}
	}
	return &s, nil
}

// Update modifies an existing student record.
func (r *Repository) Update(ctx context.Context, s Student) error {
	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		return fmt.Errorf("update student: begin tx: %w", err)
	}
	query := fmt.Sprintf(`UPDATE %s SET %s WHERE id = :id`, tableName, UpdateSetStr())

	// Key fix here too.
	result, err := tx.NamedExecContext(ctx, query, toNamedArgs(s))
	if err != nil {
		return fmt.Errorf("update student: exec update: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update student: rows affected check: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a student record by ID.
func (r *Repository) Delete(ctx context.Context, id string) error {
	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		return fmt.Errorf("delete student: tx required: %w", err)
	}
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, tableName)
	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete student: exec delete: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete student: rows affected check: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListByGrade returns all students for a specific grade.
func (r *Repository) ListByGrade(ctx context.Context, grade int32) ([]Student, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE grade = $1`, ColumnsStr(), tableName)
	var result []Student

	if tx, err := txmanager.GetTx(ctx); err == nil {
		if err := tx.SelectContext(ctx, &result, query, grade); err != nil {
			return nil, fmt.Errorf("list students by grade (tx): %w", err)
		}
		return result, nil
	}
	if err := r.db.SelectContext(ctx, &result, query, grade); err != nil {
		return nil, fmt.Errorf("list students by grade: %w", err)
	}
	return result, nil
}

// List returns all student records from the database.
func (r *Repository) List(ctx context.Context) ([]Student, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s`, ColumnsStr(), tableName)
	var students []Student
	if tx, err := txmanager.GetTx(ctx); err == nil {
		if err := tx.SelectContext(ctx, &students, query); err != nil {
			return nil, fmt.Errorf("list students (tx): %w", err)
		}
		return students, nil
	}
	if err := r.db.SelectContext(ctx, &students, query); err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	return students, nil
}

// --- helpers for NamedExec binding ---

// toNamedArgs prepares a map for sqlx.NamedExec so JSONB fields are passed as text (or nil)
// and CAST(:col AS jsonb) in SQL can parse them.
func toNamedArgs(s Student) map[string]any {
	m := map[string]any{
		"id":          s.ID,
		"first_name":  s.FirstName,
		"last_name":   s.LastName,
		"grade":       s.Grade,
		"created_at":  s.CreatedAt,
		"friends":     s.Friends,
		"middle_name": nullStringOrNil(s.MiddleName),
		"status":      nullStringOrNil(s.Status),
	}
	if len(s.HomeAddress) > 0 {
		m["home_address"] = string(s.HomeAddress)
	} else {
		m["home_address"] = nil
	}
	if len(s.CourseGrades) > 0 {
		m["course_grades"] = string(s.CourseGrades)
	} else {
		m["course_grades"] = nil
	}
	if len(s.Local) > 0 {
		m["local"] = string(s.Local)
	} else {
		m["local"] = nil
	}
	if len(s.Exchange) > 0 {
		m["exchange"] = string(s.Exchange)
	} else {
		m["exchange"] = nil
	}
	return m
}

func nullStringOrNil(ns sql.NullString) any {
	if ns.Valid {
		return ns.String
	}
	return nil
}
