package domain

import "context"

// StudentRepository defines the contract for the student persistence layer.
// It accepts and returns only domain models.
type StudentRepository interface {
	Create(ctx context.Context, s Student) (string, error)
	GetByID(ctx context.Context, id string) (*Student, error)
	Update(ctx context.Context, s Student) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]Student, error)
	ListByGrade(ctx context.Context, grade int32) ([]Student, error)
}

// StudentService defines the domain service interface.
type StudentService interface {
	Create(ctx context.Context, s Student) (string, error)
	Get(ctx context.Context, id string) (*Student, error)
	Update(ctx context.Context, s Student) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]Student, error)
	ListByGrade(ctx context.Context, grade int32) ([]Student, error)
}
