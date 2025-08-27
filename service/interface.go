// Package service contains interfaces for repository, transaction manager, and time provider used by StudentService.
package service

import (
	"context"
	"time"

	d "example.com/student-service/domain"
)

//go:generate mockgen -destination=./mocks/repository_mock.go -package=mocks example.com/student-service/service Repository
//go:generate mockgen -destination=./mocks/time_mock.go       -package=mocks example.com/student-service/service TimeProvider
//go:generate mockgen -destination=./mocks/txmanager_mock.go  -package=mocks example.com/student-service/service TxManager

// Repository defines the contract for student data storage operations.
type Repository interface {
	Create(ctx context.Context, st d.Student) (string, error)
	GetByID(ctx context.Context, id string) (*d.Student, error)
	Update(ctx context.Context, st d.Student) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]d.Student, error)
	ListByGrade(ctx context.Context, grade int32) ([]d.Student, error)
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
