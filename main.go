package main

import (
    "log"
    "net"

    "example.com/student-service/proto"
    "example.com/student-service/server"
    "google.golang.org/grpc/reflection"
    "google.golang.org/grpc"
)

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    proto.RegisterStudentServiceServer(grpcServer, server.NewStudentServer())

    reflection.Register(grpcServer)
    log.Println("gRPC server started on :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
