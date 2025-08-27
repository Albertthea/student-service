// Package app wires config, DB, tx-manager, domain service and gRPC handler together.
package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"example.com/student-service/handler"
	"example.com/student-service/internal/config"
	"example.com/student-service/internal/txmanager"
	pb "example.com/student-service/proto"
	"example.com/student-service/repository/student"
	"example.com/student-service/service"
)

// App holds the gRPC server and listener.
type App struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// realTimeProvider implements service.TimeProvider.
type realTimeProvider struct{}

// Now returns current time.
func (realTimeProvider) Now() time.Time { return time.Now() }

// realIDGenerator implements service.IDGenerator.
type realIDGenerator struct{}

// GenerateID returns a new UUID string.
func (realIDGenerator) GenerateID() string { return uuid.New().String() }

// NewApp constructs a configured gRPC server instance.
func NewApp(cfg *config.Config) (*App, error) {
	// Подключение к БД: DSN формируется конфигом.
	dsn := cfg.BuildPostgresDSN()
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect DB: %w", err)
	}

	// Инфраструктура.
	repo := student.NewRepository(db)
	txMgr := txmanager.NewManager(db)
	timeProv := realTimeProvider{}
	idGen := realIDGenerator{}

	// Доменный сервис (domain-модели на вход/выход).
	svc := service.New(newRepoAdapter(repo), txMgr, timeProv, idGen)

	// gRPC сервер и handler (вся proto<->domain конвертация живёт в handler).
	grpcServer := grpc.NewServer()
	pb.RegisterStudentServiceServer(grpcServer, handler.NewStudentServer(svc))

	// Reflection включаем только если задано в конфиге.
	if cfg.GRPC.EnableReflection {
		reflection.Register(grpcServer)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", addr, err)
	}
	log.Printf("gRPC server will start on %s", addr)

	return &App{grpcServer: grpcServer, listener: lis}, nil
}

// Run starts the gRPC server and stops it gracefully on context cancel.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := a.grpcServer.Serve(a.listener); err != nil {
			errCh <- err
		}
	}()
	select {
	case <-ctx.Done():
		log.Println("context cancelled: graceful stop")
		a.grpcServer.GracefulStop()
		return nil
	case err := <-errCh:
		return fmt.Errorf("grpc serve: %w", err)
	}
}
