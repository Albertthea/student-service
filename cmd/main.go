// Package main starts the gRPC student-service server.
package main

import (
	"fmt"
	"log"
	"net"

	"example.com/student-service/internal/config"
	"example.com/student-service/proto"
	"example.com/student-service/repository/student"
	"example.com/student-service/service"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // import PostgreSQL driver
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=studentdb sslmode=disable",
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.Port,
		cfg.DBLogin,
		cfg.DBPassword,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
	}()

	repo := student.NewRepository(db)
	studentService := service.NewStudentServer(repo)

	listenAddress := fmt.Sprintf(":%d", cfg.Server.Port)
	lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", listenAddress, err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterStudentServiceServer(grpcServer, studentService)
	reflection.Register(grpcServer)

	log.Printf("gRPC server started on %s", listenAddress)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
