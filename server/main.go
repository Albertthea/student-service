package main

import (
    "log"
    "net"
    "student-service/proto"
    "student-service/server"
    "google.golang.org/grpc"
)

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    proto.RegisterStudentServiceServer(grpcServer, server.NewStudentServer())

    log.Println("gRPC server started on :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
