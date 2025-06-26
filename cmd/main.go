// Package main starts the gRPC student-service server.
package main

import (
	"log"
	"net"

	"example.com/student-service/proto"
	"example.com/student-service/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

<<<<<<< Updated upstream
    grpcServer := grpc.NewServer()
    proto.RegisterStudentServiceServer(grpcServer, server.NewStudentServer())
=======
	grpcServer := grpc.NewServer()
	proto.RegisterStudentServiceServer(grpcServer, service.NewStudentServer())
>>>>>>> Stashed changes

	reflection.Register(grpcServer)
	log.Println("gRPC server started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
