// Package service implements the student gRPC service logic.
package service

import (
	"context"
	"sync"
	"time"

	"example.com/student-service/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Student represents a student record.
type Student struct {
	ID        string
	FirstName string
	LastName  string
	Grade     int32
	CreatedAt time.Time
}

// Store holds the in-memory database.
type Store struct {
	mu       sync.RWMutex
	students map[string]Student
}

// StudentServer implements the StudentServiceServer gRPC interface.
type StudentServer struct {
	proto.UnimplementedStudentServiceServer
	store *Store
}

// NewStudentServer creates a new instance of StudentServer.
func NewStudentServer() *StudentServer {
	return &StudentServer{
		store: &Store{
			students: make(map[string]Student),
		},
	}
}

// CreateStudent handles a gRPC request to create a new student.
func (s *StudentServer) CreateStudent(_ context.Context, req *proto.CreateStudentRequest) (*proto.CreateStudentResponse, error) {
	id := uuid.New().String()
	now := time.Now()

	s.store.mu.Lock()
	s.store.students[id] = Student{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Grade:     req.Grade,
		CreatedAt: now,
	}
	s.store.mu.Unlock()

	return &proto.CreateStudentResponse{Id: id}, nil
}

// GetStudent handles a gRPC request to retrieve a student by ID.
func (s *StudentServer) GetStudent(_ context.Context, req *proto.GetStudentRequest) (*proto.GetStudentResponse, error) {
	s.store.mu.RLock()
	student, ok := s.store.students[req.Id]
	s.store.mu.RUnlock()

	if !ok {
		return nil, status.Errorf(codes.NotFound, "student with id %s not found", req.Id)
	}

	return &proto.GetStudentResponse{
		Student: &proto.Student{
			Id:        student.ID,
			FirstName: student.FirstName,
			LastName:  student.LastName,
			Grade:     student.Grade,
			CreatedAt: timestamppb.New(student.CreatedAt),
		},
	}, nil
}

// UpdateStudent handles a gRPC request to update an existing student's data.
func (s *StudentServer) UpdateStudent(_ context.Context, req *proto.UpdateStudentRequest) (*emptypb.Empty, error) {
	student := req.Student

	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	stored, ok := s.store.students[student.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "student with id %s not found", student.Id)
	}

	if student.CreatedAt != nil {
		updatedCreatedAt := student.CreatedAt.AsTime()
		if !updatedCreatedAt.Equal(stored.CreatedAt) {
			return nil, status.Errorf(codes.InvalidArgument, "created_at field cannot be modified")
		}
	}

	if student.Grade < stored.Grade {
		return nil, status.Errorf(codes.FailedPrecondition, "grade cannot be decreased")
	}

	stored.FirstName = student.FirstName
	stored.LastName = student.LastName
	stored.Grade = student.Grade

	s.store.students[student.Id] = stored

	return &emptypb.Empty{}, nil
}

// DeleteStudent handles a gRPC request to delete a student by ID.
func (s *StudentServer) DeleteStudent(_ context.Context, req *proto.DeleteStudentRequest) (*emptypb.Empty, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	student := req.Id

	if _, ok := s.store.students[student]; !ok {
		return nil, status.Errorf(codes.NotFound, "student not found")
	}

	delete(s.store.students, student)
	return &emptypb.Empty{}, nil
}

// ListStudents handles a gRPC request to return all students.
func (s *StudentServer) ListStudents(_ context.Context, req *proto.ListStudentsRequest) (*proto.ListStudentsResponse, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	var result []*proto.Student

	for _, student := range s.store.students {
		if student.Grade == req.Grade {
			result = append(result, &proto.Student{
				Id:        student.ID,
				FirstName: student.FirstName,
				LastName:  student.LastName,
				Grade:     student.Grade,
				CreatedAt: timestamppb.New(student.CreatedAt),
			})
		}
	}

	return &proto.ListStudentsResponse{Students: result}, nil
}
