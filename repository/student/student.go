// Package student provides the data structures and repository methods
// for managing student records.
package student

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

var (
	// ErrNotFound is returned when a student record is not found in the database.
	ErrNotFound = errors.New("student not found")
	// ErrAlreadyExists is returned when trying to create a student that already exists.
	ErrAlreadyExists = errors.New("student already exists")
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

	query := fmt.Sprintf(
		`INSERT INTO students (%s) VALUES %s`,
		ColumnsStr(),
		Placeholders(len(Columns)),
	)

	_, err = tx.ExecContext(ctx, query,
		s.ID, s.FirstName, s.LastName, s.Grade, s.CreatedAt,
	)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("tx.Rollback error: %v", rbErr)
		}
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return s.ID, nil
}

// GetByID retrieves a student by their ID.
func (r *Repository) GetByID(ctx context.Context, id string) (Student, error) {
	var s Student
	query := fmt.Sprintf(`SELECT %s FROM students WHERE id = $1`, ColumnsStr())
	err := r.db.QueryRowContext(ctx, query, id).
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

	query := `UPDATE students SET first_name = $1, last_name = $2, grade = $3 WHERE id = $4`

	result, err := tx.ExecContext(ctx, query,
		s.FirstName, s.LastName, s.Grade, s.ID,
	)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("tx.Rollback error: %v", rbErr)
		}
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("tx.Rollback error: %v", rbErr)
		}
		return err
	}
	if rowsAffected == 0 {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("tx.Rollback error: %v", rbErr)
		}
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

	query := `DELETE FROM students WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("tx.Rollback error: %v", rbErr)
		}
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("tx.Rollback error: %v", rbErr)
		}
		return err
	}
	if rowsAffected == 0 {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("tx.Rollback error: %v", rbErr)
		}
		return ErrNotFound
	}
	return tx.Commit()
}

// ListByGrade returns all students for a specific grade.
func (r *Repository) ListByGrade(ctx context.Context, grade int32) ([]Student, error) {
	query := fmt.Sprintf(`SELECT %s FROM students WHERE grade = $1`, ColumnsStr())
	rows, err := r.db.QueryContext(ctx, query, grade)
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
