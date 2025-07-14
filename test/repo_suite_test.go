package test

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"example.com/student-service/internal/txmanager"
	"example.com/student-service/repository/student"

	_ "github.com/lib/pq"
)

//go:embed migrations/00000_initial.sql
var migrationSQL string

type RepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	container *postgres.PostgresContainer
	db        *sqlx.DB
	repo      *student.Repository
}

func (s *RepoTestSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := postgres.Run(
		s.ctx,
		"docker.io/postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithStartupTimeout(30*time.Second),
		),
	)
	s.Require().NoError(err)
	s.container = container

	dsn, err := container.ConnectionString(s.ctx, "sslmode=disable")
	s.Require().NoError(err)

	db, err := sqlx.Open("postgres", dsn)
	s.Require().NoError(err)
	s.Require().NoError(waitForDB(db))

	_, err = db.Exec(migrationSQL)
	s.Require().NoError(err)

	s.db = db
	s.repo = student.NewRepository(db)
}

func (s *RepoTestSuite) TestCreateGetUpdateDeleteStudent() {
	txMgr := txmanager.NewManager(s.repo.DB())

	st := student.Student{
		ID:        uuid.New().String(),
		FirstName: "Ivan",
		LastName:  "Ivanov",
		Grade:     5,
	}

	var id string
	err := txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
		var err error
		id, err = s.repo.Create(ctx, st)
		return err
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(id)

	// Get after transaction committed
	got, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal("Ivan", got.FirstName)
	s.Equal("Ivanov", got.LastName)
	s.Equal(int32(5), got.Grade)

	// Update inside transaction
	got.LastName = "Petrov"
	got.Grade = 6
	err = txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
		return s.repo.Update(ctx, *got)
	})
	s.Require().NoError(err)

	// Get updated
	updated, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal("Petrov", updated.LastName)
	s.Equal(int32(6), updated.Grade)

	// Delete inside transaction
	err = txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
		return s.repo.Delete(ctx, id)
	})
	s.Require().NoError(err)

	// Verify deletion
	_, err = s.repo.GetByID(s.ctx, id)
	s.Require().Error(err)
	s.ErrorIs(err, student.ErrNotFound)
}

func (s *RepoTestSuite) TestListAndListByGrade() {
	txMgr := txmanager.NewManager(s.repo.DB())

	_, _, err := s.container.Exec(s.ctx, []string{
		"psql", "-U", "postgres", "-d", "testdb", "-c", "DELETE FROM students;",
	})
	s.Require().NoError(err)

	grades := []int32{1, 2, 2, 3}
	for i, g := range grades {
		st := student.Student{
			ID:        uuid.New().String(),
			FirstName: fmt.Sprintf("Name%d", i),
			LastName:  "Test",
			Grade:     g,
		}
		err := txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
			_, err := s.repo.Create(ctx, st)
			return err
		})
		s.Require().NoError(err)
	}

	all, err := s.repo.List(s.ctx)
	s.Require().NoError(err)
	s.Len(all, len(grades))

	gradeTwo, err := s.repo.ListByGrade(s.ctx, 2)
	s.Require().NoError(err)
	s.Len(gradeTwo, 2)

	none, err := s.repo.ListByGrade(s.ctx, 99)
	s.Require().NoError(err)
	s.Empty(none)
}

func (s *RepoTestSuite) TestTxManager_SuccessAndRollback() {
	txMgr := txmanager.NewManager(s.repo.DB())

	_, err := s.repo.DB().ExecContext(s.ctx, "DELETE FROM dummy;")
	s.Require().NoError(err)

	// successful
	err = txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
		tx, err := txmanager.GetTx(ctx)
		s.Require().NoError(err)
		_, err = tx.ExecContext(ctx, `INSERT INTO dummy (id, value) VALUES ('id1', 'value1')`)
		return err
	})
	s.Require().NoError(err)

	var value string
	err = s.repo.DB().GetContext(s.ctx, &value, "SELECT value FROM dummy WHERE id = $1", "id1")
	s.Require().NoError(err)
	s.Equal("value1", value)

	// rollback
	err = txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
		tx, err := txmanager.GetTx(ctx)
		s.Require().NoError(err)
		_, err = tx.ExecContext(ctx, `INSERT INTO dummy (id, value) VALUES ('id2', 'value2')`)
		s.Require().NoError(err)
		return errors.New("simulate rollback")
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "simulate rollback")

	var count int
	err = s.repo.DB().GetContext(s.ctx, &count, "SELECT COUNT(*) FROM dummy WHERE id = $1", "id2")
	s.Require().NoError(err)
	s.Equal(0, count)
}

func (s *RepoTestSuite) TearDownSuite() {
	if err := s.db.Close(); err != nil {
		s.T().Logf("Error closing DB: %v", err)
	}
	err := s.container.Terminate(s.ctx)
	s.Require().NoError(err)
}

func TestRepoTestSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

func waitForDB(db *sqlx.DB) error {
	const maxAttempts = 10
	const delay = time.Second
	for i := 0; i < maxAttempts; i++ {
		if err := db.Ping(); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return errors.New("database not reachable after waiting")
}
