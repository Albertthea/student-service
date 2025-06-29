package student

import (
	"context"
	"database/sql"
	"time"
)

type Student struct {
	ID         string
	FirstName  string
	LastName   string
	Grade      int32
	CreatedAt  time.Time
}

type StudentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) Create(ctx context.Context, s Student) (string, error) {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO students (id, first_name, last_name, grade, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		s.ID, s.FirstName, s.LastName, s.Grade, s.CreatedAt,
	)
	return s.ID, err
}

func (r *StudentRepository) GetByID(ctx context.Context, id string) (Student, error) {
	var s Student
	err := r.db.QueryRowContext(ctx,
		`SELECT id, first_name, last_name, grade, created_at FROM students WHERE id = $1`, id).
		Scan(&s.ID, &s.FirstName, &s.LastName, &s.Grade, &s.CreatedAt)
	return s, err
}

func (r *StudentRepository) Update(ctx context.Context, s Student) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE students SET first_name = $1, last_name = $2, grade = $3 WHERE id = $4`,
		s.FirstName, s.LastName, s.Grade, s.ID,
	)
	return err
}

func (r *StudentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM students WHERE id = $1`, id)
	return err
}

func (r *StudentRepository) ListByGrade(ctx context.Context, grade int32) ([]Student, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, first_name, last_name, grade, created_at FROM students WHERE grade = $1`, grade)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
