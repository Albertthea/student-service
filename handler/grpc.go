package handler

import (
	d "example.com/student-service/domain"
	"example.com/student-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Register registers the StudentService gRPC server with the given gRPC server instance.
func Register(s *grpc.Server, srv proto.StudentServiceServer) {
	proto.RegisterStudentServiceServer(s, srv)
}

// ToStatus converts a domain/service error into a gRPC status error.
func ToStatus(err error) *status.Status {
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok {
		return st
	}

	switch err {
	case d.ErrNotFound:
		return status.New(codes.NotFound, err.Error())
	case d.ErrAlreadyExists:
		return status.New(codes.AlreadyExists, err.Error())
	case d.ErrInvalidArgument, d.ErrCreatedAtImmutable, d.ErrDetailsRequired:
		return status.New(codes.InvalidArgument, err.Error())
	case d.ErrGradeDecrease:
		return status.New(codes.FailedPrecondition, err.Error())
	default:
		return status.New(codes.Internal, err.Error())
	}
}
