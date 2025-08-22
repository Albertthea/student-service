// Package handler contains gRPC server registration and error translation helpers.
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
func ToStatus(err error) error {
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok {
		return st.Err()
	}

	switch err {
	case d.ErrNotFound:
		return status.Error(codes.NotFound, err.Error())
	case d.ErrCreatedAtImmutable:
		return status.Error(codes.InvalidArgument, err.Error())
	case d.ErrGradeDecrease:
		return status.Error(codes.FailedPrecondition, err.Error())
	case d.ErrDetailsRequired:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
