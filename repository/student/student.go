// Package student provides the data structures and repository methods
// for managing student records.
package student

import (
	"context"
	"database/sql"
	"log"
	"time"
	"errors"
)

var (
	ErrNotFound = errors.New("student not found")
	ErrAlreadyExists  = errors.New("student already exists")
)

// Student represents a student record in the database.
type Student struct {
	ID        string
	FirstName string
	LastName  string
	Grade     int32
	CreatedAt time.Time
}

// Repository manages operations with students in the DB.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Repository with the given DB connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new student record into the database.
func (r *Repository) Create(ctx context.Context, s Student) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}

	_, err := tx.ExecContext(ctx,
		`INSERT INTO students (id, first_name, last_name, grade, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		s.ID, s.FirstName, s.LastName, s.Grade, s.CreatedAt,
	)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return s.ID, err
}

// GetByID retrieves a student by their ID.
func (r *Repository) GetByID(ctx context.Context, id string) (Student, error) {
	var s Student
	err := r.db.QueryRowContext(ctx,
		`SELECT id, first_name, last_name, grade, created_at FROM students WHERE id = $1`, id).
		Scan(&s.ID, &s.FirstName, &s.LastName, &s.Grade, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Student{}, ErrNotFound
		}
		return Student{}, err
	}
	return s, nil
}

// Update modifies an existing student record.
func (r *Repository) Update(ctx context.Context, s Student) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx,
		`UPDATE students SET first_name = $1, last_name = $2, grade = $3 WHERE id = $4`,
		s.FirstName, s.LastName, s.Grade, s.ID,
	)

	if err != nil {
		tx.Rollback()
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return ErrNotFound
	}
	return tx.Commit()
}

// Delete removes a student record by ID.
func (r *Repository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	
	result, err := tx.ExecContext(ctx,
		`DELETE FROM students WHERE id = $1`, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return ErrNotFound
	}
	return tx.Commit()
}

// ListByGrade returns all students for a specific grade.
func (r *Repository) ListByGrade(ctx context.Context, grade int32) ([]Student, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, first_name, last_name, grade, created_at FROM students WHERE grade = $1`, grade)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var result []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.Grade, &s.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
