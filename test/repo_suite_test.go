package test

import (
	"context"
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
		postgres.WithInitScripts("migrations/init.sql"),
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

	s.repo = student.NewRepository(db)
}

func (s *RepoTestSuite) TearDownSuite() {
	err := s.container.Terminate(s.ctx)
	s.Require().NoError(err)
}

func TestRepoTestSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}
