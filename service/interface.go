// Package service contains interfaces for repository, transaction manager, and time provider used by StudentService.
package service

import (
	"context"
	"time"

	"example.com/student-service/repository/student"
)

// Repository defines the contract for student data storage operations.
type Repository interface {
	Create(ctx context.Context, st student.Student) (string, error)
	GetByID(ctx context.Context, id string) (*student.Student, error)
	Update(ctx context.Context, st student.Student) error
	Delete(ctx context.Context, id string) error
	ListByGrade(ctx context.Context, grade int32) ([]student.Student, error)
}

// TimeProvider provides the current time. Useful for testing time-dependent logic.
type TimeProvider interface {
	Now() time.Time
}

// TxManager defines a transaction manager capable of executing functions within a transactional context.
type TxManager interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}

// IDGenerator generates unique IDs for new entities.
type IDGenerator interface {
	GenerateID() string
}
