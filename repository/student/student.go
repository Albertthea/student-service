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
	ID        string    `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Grade     int32     `db:"grade"`
	CreatedAt time.Time `db:"created_at"`
}

// Repository manages operations with students in the DB.
type Repository struct {
	db *sqlx.DB
}

// DB returns the underlying sqlx.DB instance.
func (r *Repository) DB() *sqlx.DB {
	return r.db
}

// NewRepository creates a new Repository with the given DB connection.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new student record into the database.
func (r *Repository) Create(ctx context.Context, s Student) (string, error) {
	tx, err := txmanager.GetTx(ctx)
	if err != nil {
		return "", fmt.Errorf("create student: tx required: %w", err)
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s`, tableName, ColumnsStr(), NamedPlaceholders())

	_, err = tx.NamedExecContext(ctx, query, &s)

	if err != nil {
		return "", fmt.Errorf("create student: insert: %w", err)
	}

	return s.ID, nil
}

// GetByID retrieves a student by their ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*Student, error) {
	var s Student
	query := fmt.Sprintf(`SELECT id, first_name, last_name, grade, created_at FROM %s WHERE id = $1`, tableName)

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

	result, err := tx.NamedExecContext(ctx, query, &s)
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
	query := fmt.Sprintf(`SELECT id, first_name, last_name, grade, created_at FROM %s WHERE grade = $1`, tableName)
	var result []Student
	err := r.db.SelectContext(ctx, &result, query, grade)
	if err != nil {
		return nil, fmt.Errorf("list students by grade: %w", err)
	}
	return result, nil
}
