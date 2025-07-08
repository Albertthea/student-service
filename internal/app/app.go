// Package app contains the implementation for running and configuring the student-service application.
package app

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"example.com/student-service/internal/config"
	"example.com/student-service/internal/txmanager"
	"example.com/student-service/proto"
	"example.com/student-service/repository/student"
	"example.com/student-service/service"
)

// App represents the gRPC server and listener for the student-service.
type App struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// realTimeProvider implements the TimeProvider interface, returning the current time.
type realTimeProvider struct{}

// Now returns the current local time.
func (realTimeProvider) Now() time.Time {
	return time.Now()
}

// realIDGenerator implements the IDGenerator interface, generating UUIDs.
type realIDGenerator struct{}

// GenerateID generates a new UUID string.
func (realIDGenerator) GenerateID() string {
	return uuid.New().String()
}

// NewApp creates a new App instance using the provided configuration.
func NewApp(cfg *config.Config) (*App, error) {
	dbLogin := os.Getenv(cfg.PostgreSQL.Authorisation.Env.LoginEnv)
	dbPassword := os.Getenv(cfg.PostgreSQL.Authorisation.Env.PasswordEnv)

	if dbLogin == "" || dbPassword == "" {
		return nil, fmt.Errorf("missing DB credentials in env variables: %s, %s",
			cfg.PostgreSQL.Authorisation.Env.LoginEnv,
			cfg.PostgreSQL.Authorisation.Env.PasswordEnv)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=studentdb sslmode=disable",
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.Port,
		dbLogin,
		dbPassword,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	repo := student.NewRepository(db)
	timeProvider := realTimeProvider{}
	idGen := realIDGenerator{}
	txManager := txmanager.NewManager(db)

	server := service.NewStudentServer(repo, timeProvider, txManager, idGen)
	grpcServer := grpc.NewServer()
	proto.RegisterStudentServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	listenAddress := fmt.Sprintf(":%d", cfg.Server.Port)
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", listenAddress, err)
	}

	log.Printf("gRPC server will start on %s", listenAddress)

	return &App{
		grpcServer: grpcServer,
		listener:   listener,
	}, nil
}

// Run starts the gRPC server and listens for incoming connections.
func (a *App) Run() error {
	return a.grpcServer.Serve(a.listener)
}
