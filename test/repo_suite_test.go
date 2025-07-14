package test

import (
	"context"
	_ "embed"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"example.com/student-service/repository/student"
	"example.com/student-service/service"
)

//go:embed migrations/00000_initial.sql
var migrationSQL string

type RepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	container *postgres.PostgresContainer
	repo      service.Repository
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

	_, err = db.Exec(migrationSQL)
	s.Require().NoError(err)

	s.repo = student.NewRepository(db)
}

func (s *RepoTestSuite) TestCreateGetUpdateDeleteStudent() {
	// Create a new student record
	st := student.Student{
		FirstName: "Ivan",
		LastName:  "Ivanov",
		Grade:     5,
	}
	id, err := s.repo.Create(s.ctx, st)
	s.Require().NoError(err)
	s.Require().NotEmpty(id)

	// Retrieve the student by ID and verify the fields
	got, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal("Ivan", got.FirstName)
	s.Equal("Ivanov", got.LastName)
	s.Equal(int32(5), got.Grade)

	// Update student's last name and grade
	got.LastName = "Petrov"
	got.Grade = 6
	err = s.repo.Update(s.ctx, *got)
	s.Require().NoError(err)

	// Retrieve again to confirm updates persisted
	updated, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal("Petrov", updated.LastName)
	s.Equal(int32(6), updated.Grade)

	// Delete the student record
	err = s.repo.Delete(s.ctx, id)
	s.Require().NoError(err)

	// Verify that the student no longer exists
	deleted, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Nil(deleted)
}

func (s *RepoTestSuite) TestListAndListByGrade() {
	// Clean table
	_, _, err := s.container.Exec(s.ctx, []string{"psql", "-U", "postgres", "-d", "testdb", "-c", "DELETE FROM students;"})
	s.Require().NoError(err)

	// Create students with different grade
	grades := []int32{1, 2, 2, 3}
	for i, g := range grades {
		st := student.Student{
			FirstName: fmt.Sprintf("Name%d", i),
			LastName:  "Test",
			Grade:     g,
		}
		_, err := s.repo.Create(s.ctx, st)
		s.Require().NoError(err)
	}

	// Check that list return all
	all, err := s.repo.List(s.ctx)
	s.Require().NoError(err)
	s.Len(all, len(grades))

	// Should be 2 students
	gradeTwo, err := s.repo.ListByGrade(s.ctx, 2)
	s.Require().NoError(err)
	s.Len(gradeTwo, 2)

	// List should be empty
	none, err := s.repo.ListByGrade(s.ctx, 99)
	s.Require().NoError(err)
	s.Empty(none)
}

func (s *RepoTestSuite) TearDownSuite() {
	err := s.container.Terminate(s.ctx)
	s.Require().NoError(err)
}

func TestRepoTestSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}
