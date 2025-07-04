// Package main starts the gRPC student-service server.
package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"example.com/student-service/internal/config"
	"example.com/student-service/proto"
	"example.com/student-service/repository/student"
	"example.com/student-service/service"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // import PostgreSQL driver
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ConfigPath defines the path to the YAML configuration file.
const ConfigPath = "config.yaml"

func main() {
	cfg, err := config.LoadConfig(ConfigPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dbLogin := os.Getenv(cfg.PostgreSQL.Authorisation.Env.LoginEnv)
	dbPassword := os.Getenv(cfg.PostgreSQL.Authorisation.Env.PasswordEnv)

	if dbLogin == "" || dbPassword == "" {
		log.Fatalf("missing DB credentials in env variables: %s, %s",
			cfg.PostgreSQL.Authorisation.Env.LoginEnv,
			cfg.PostgreSQL.Authorisation.Env.PasswordEnv)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=studentdb sslmode=disable",
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.Port,
		dbLogin,
		dbPassword,
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
