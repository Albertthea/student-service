// Package service implements the student gRPC service logic.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"example.com/student-service/internal/txmanager"
	"example.com/student-service/proto"
	"example.com/student-service/repository/student"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// StudentServer implements the StudentServiceServer gRPC interface.
type StudentServer struct {
	proto.UnimplementedStudentServiceServer
	repo *student.Repository
}

// NewStudentServer creates a new instance of StudentServer.
func NewStudentServer(repo *student.Repository) *StudentServer {
	return &StudentServer{
		repo: repo,
	}
}

// CreateStudent handles a gRPC request to create a new student.
func (s *StudentServer) CreateStudent(ctx context.Context, req *proto.CreateStudentRequest) (*proto.CreateStudentResponse, error) {
	id := uuid.New().String()
	now := time.Now()

	st := student.Student{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Grade:     req.Grade,
		CreatedAt: now,
	}
	err := txmanager.WithTransaction(ctx, s.repo.DB(), func(txCtx context.Context) error {
		_, err := s.repo.Create(txCtx, st)
		return err
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create student: %v", err)
	}

	return &proto.CreateStudentResponse{Id: id}, nil
}

// GetStudent handles a gRPC request to retrieve a student by ID.
func (s *StudentServer) GetStudent(ctx context.Context, req *proto.GetStudentRequest) (*proto.GetStudentResponse, error) {
	st, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, student.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "student with id %s not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to get student: %v", err)
	}

	return &proto.GetStudentResponse{
		Student: &proto.Student{
			Id:        st.ID,
			FirstName: st.FirstName,
			LastName:  st.LastName,
			Grade:     st.Grade,
			CreatedAt: timestamppb.New(st.CreatedAt),
		},
	}, nil
}

// UpdateStudent handles a gRPC request to update an existing student's data.
func (s *StudentServer) UpdateStudent(ctx context.Context, req *proto.UpdateStudentRequest) (*emptypb.Empty, error) {
	err := txmanager.WithTransaction(ctx, s.repo.DB(), func(txCtx context.Context) error {
		existing, err := s.repo.GetByID(txCtx, req.Student.Id)
		if err != nil {
			if errors.Is(err, student.ErrNotFound) {
				return fmt.Errorf("student not found")
			}
			return fmt.Errorf("failed to fetch student: %v", err)
		}

		if req.Student.CreatedAt != nil && !req.Student.CreatedAt.AsTime().Equal(existing.CreatedAt) {
			return fmt.Errorf("created_at field cannot be modified")
		}

		if req.Student.Grade < existing.Grade {
			return fmt.Errorf("grade cannot be decreased")
		}

		updated := student.Student{
			ID:        req.Student.Id,
			FirstName: req.Student.FirstName,
			LastName:  req.Student.LastName,
			Grade:     req.Student.Grade,
			CreatedAt: existing.CreatedAt,
		}

		return s.repo.Update(txCtx, updated)
	})

	if err != nil {
		switch err.Error() {
		case "student not found":
			return nil, status.Errorf(codes.NotFound, err.Error())
		case "created_at field cannot be modified":
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		case "grade cannot be decreased":
			return nil, status.Errorf(codes.FailedPrecondition, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "failed to update student: %v", err)
		}
	}

	return &emptypb.Empty{}, nil
}

// DeleteStudent handles a gRPC request to delete a student by ID.
func (s *StudentServer) DeleteStudent(ctx context.Context, req *proto.DeleteStudentRequest) (*emptypb.Empty, error) {
	err := txmanager.WithTransaction(ctx, s.repo.DB(), func(txCtx context.Context) error {
		return s.repo.Delete(txCtx, req.Id)
	})
	if err != nil {
		if errors.Is(err, student.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "student not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete student: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// ListStudents handles a gRPC request to return all students.
func (s *StudentServer) ListStudents(ctx context.Context, req *proto.ListStudentsRequest) (*proto.ListStudentsResponse, error) {
	students, err := s.repo.ListByGrade(ctx, req.Grade)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list students: %v", err)
	}

	var result []*proto.Student

	for _, st := range students {
		result = append(result, &proto.Student{
			Id:        st.ID,
			FirstName: st.FirstName,
			LastName:  st.LastName,
			Grade:     st.Grade,
			CreatedAt: timestamppb.New(st.CreatedAt),
		})
	}

	return &proto.ListStudentsResponse{Students: result}, nil
}
