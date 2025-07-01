// Package main starts the gRPC student-service server.
package main

import (
	"database/sql"
	"log"
	"net"

	"example.com/student-service/proto"
	"example.com/student-service/repository/student"
	"example.com/student-service/service"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	dsn := "host=localhost port=5434 user=student password=111111 dbname=studentdb sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatalf("db is unreachable: %v", err)
	}

	repo := student.NewRepository(db)
	svc := service.NewStudentServer(repo)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterStudentServiceServer(grpcServer, svc)
	reflection.Register(grpcServer)

	log.Println("gRPC server started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
