// Package service implements domain-level student use cases.
package service

import (
	"context"
	"fmt"
	"time"

	d "example.com/student-service/domain"
)

// Service implements business logic for students (no proto/gRPC).
type Service struct {
	repo         Repository
	tx           TxManager
	timeProvider TimeProvider
	idGen        IDGenerator
}

// New creates a new Service.
func New(repo Repository, tx TxManager, tp TimeProvider, idg IDGenerator) *Service {
	return &Service{repo: repo, tx: tx, timeProvider: tp, idGen: idg}
}

// Create stores a new student in the repository.
func (s *Service) Create(ctx context.Context, st d.Student) (string, error) {
	if st.ID == "" {
		st.ID = s.idGen.GenerateID()
	}
	if st.CreatedAt.IsZero() {
		st.CreatedAt = s.timeProvider.Now().UTC().Truncate(time.Microsecond)
	}
	if err := st.ValidateCreate(); err != nil {
		return "", err
	}
	var id string
	if err := s.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		id, err = s.repo.Create(txCtx, st)
		return err
	}); err != nil {
		return "", err
	}
	return id, nil
}

// Get retrieves a student by ID from the repository.
func (s *Service) Get(ctx context.Context, id string) (*d.Student, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", d.ErrInvalidArgument)
	}
	return s.repo.GetByID(ctx, id)
}

// Update modifies an existing student in the repository.
func (s *Service) Update(ctx context.Context, upd d.Student) error {
	if upd.ID == "" {
		return fmt.Errorf("%w: id is required", d.ErrInvalidArgument)
	}
	return s.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		cur, err := s.repo.GetByID(txCtx, upd.ID)
		if err != nil {
			return err
		}
		if upd.CreatedAt.IsZero() {
			upd.CreatedAt = cur.CreatedAt
		}
		if err := cur.CanUpdate(upd); err != nil {
			return err
		}
		return s.repo.Update(txCtx, upd)
	})
}

// Delete removes a student by ID from the repository.
func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: id is required", d.ErrInvalidArgument)
	}
	return s.tx.WithTransaction(ctx, func(txCtx context.Context) error {
		return s.repo.Delete(txCtx, id)
	})
}

// List returns all students from the repository.
func (s *Service) List(ctx context.Context) ([]d.Student, error) {
	return s.repo.List(ctx)
}

// ListByGrade returns all students with the specified grade.
func (s *Service) ListByGrade(ctx context.Context, grade int32) ([]d.Student, error) {
	if grade <= 0 {
		return s.repo.List(ctx)
	}
	return s.repo.ListByGrade(ctx, grade)
}
