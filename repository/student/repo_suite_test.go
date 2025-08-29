package student_test

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"testing"
	"time"

	"example.com/student-service/internal/txmanager"
	"example.com/student-service/repository/student"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type RepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	container *postgres.PostgresContainer
	db        *sqlx.DB
	repo      *student.Repository
}

// SetupSuite starts a Postgres container, applies goose migrations, and initializes the repository.
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

	// Apply only Up sections from embedded migrations via goose.
	goose.SetBaseFS(migrationsFS)
	s.Require().NoError(goose.SetDialect("postgres"))
	s.Require().NoError(goose.Up(db.DB, "migrations"))

	// Sanity check: the table must exist after migrations.
	var exists bool
	err = db.Get(&exists, `
		SELECT EXISTS (
		  SELECT 1
		  FROM information_schema.tables
		  WHERE table_schema = 'public' AND table_name = 'students'
		)`)
	s.Require().NoError(err)
	s.Require().True(exists, "table 'students' must exist after migrations")

	s.db = db
	s.repo = student.NewRepository(db)
}

// TestCreateGetUpdateDeleteStudent verifies CRUD operations on the repository.
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

	got, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal("Ivan", got.FirstName)
	s.Equal("Ivanov", got.LastName)
	s.Equal(int32(5), got.Grade)

	got.LastName = "Petrov"
	got.Grade = 6
	err = txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
		return s.repo.Update(ctx, *got)
	})
	s.Require().NoError(err)

	updated, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal("Petrov", updated.LastName)
	s.Equal(int32(6), updated.Grade)

	err = txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
		return s.repo.Delete(ctx, id)
	})
	s.Require().NoError(err)

	_, err = s.repo.GetByID(s.ctx, id)
	s.Require().Error(err)
	s.ErrorIs(err, student.ErrNotFound)
}

// TestListAndListByGrade verifies listing all students and filtering by grade.
func (s *RepoTestSuite) TestListAndListByGrade() {
	txMgr := txmanager.NewManager(s.repo.DB())

	_, err := s.db.ExecContext(s.ctx, `DELETE FROM students`)
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

// TearDownSuite closes DB and stops the container.
func (s *RepoTestSuite) TearDownSuite() {
	if err := s.db.Close(); err != nil {
		s.T().Logf("error closing DB: %v", err)
	}
	err := s.container.Terminate(s.ctx)
	s.Require().NoError(err)
}

// TestRepoTestSuite runs the repository test suite.
func TestRepoTestSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

// waitForDB polls the DB until it responds to Ping.
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
