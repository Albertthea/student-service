// Package service implements the student gRPC service logic.
package service

import (
	"context"
	"errors"
	"time"

	d "example.com/student-service/domain"
	"example.com/student-service/handler"
	"example.com/student-service/mapper"
	"example.com/student-service/proto"
	repo "example.com/student-service/repository/student"
	"google.golang.org/protobuf/types/known/emptypb"
)

// StudentServer implements the StudentServiceServer gRPC interface.
type StudentServer struct {
	proto.UnimplementedStudentServiceServer
	repo         Repository
	timeProvider TimeProvider
	txManager    TxManager
	idGenerator  IDGenerator
}

// NewStudentServer creates a new instance of StudentServer.
func NewStudentServer(repo Repository, timeProvider TimeProvider, txManager TxManager, idGen IDGenerator) *StudentServer {
	return &StudentServer{
		repo:         repo,
		timeProvider: timeProvider,
		txManager:    txManager,
		idGenerator:  idGen,
	}
}

// CreateStudent handles a gRPC request to create a new student.
func (s *StudentServer) CreateStudent(ctx context.Context, req *proto.CreateStudentRequest) (*proto.CreateStudentResponse, error) {
	id := s.idGenerator.GenerateID()
	now := s.timeProvider.Now()

	entity, err := mapper.CreateReqToRepo(id, now, req)
	if err != nil {
		return nil, handler.ToStatus(err).Err()
	}

	if err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		_, err := s.repo.Create(txCtx, entity)
		return err
	}); err != nil {
		return nil, handler.ToStatus(err).Err()
	}

	return &proto.CreateStudentResponse{Id: id}, nil
}

// GetStudent handles a gRPC request to retrieve a student by ID.
func (s *StudentServer) GetStudent(ctx context.Context, req *proto.GetStudentRequest) (*proto.GetStudentResponse, error) {
	st, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, handler.ToStatus(d.ErrNotFound).Err()
		}
		return nil, handler.ToStatus(err).Err()
	}
	return &proto.GetStudentResponse{Student: mapper.RepoToProto(st)}, nil
}

// UpdateStudent handles a gRPC request to update an existing student's data.
func (s *StudentServer) UpdateStudent(ctx context.Context, req *proto.UpdateStudentRequest) (*emptypb.Empty, error) {
	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		current, err := s.repo.GetByID(txCtx, req.GetStudent().GetId())
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return d.ErrNotFound
			}
			return err
		}

		curDom := d.Student{
			ID:        current.ID,
			Grade:     current.Grade,
			CreatedAt: current.CreatedAt,
		}
		var newCreatedAt time.Time
		if ts := req.GetStudent().GetCreatedAt(); ts != nil {
			newCreatedAt = ts.AsTime()
		} else {
			newCreatedAt = current.CreatedAt
		}
		newDom := d.Student{
			ID:        req.GetStudent().GetId(),
			Grade:     req.GetStudent().GetGrade(),
			CreatedAt: newCreatedAt,
		}
		if err := curDom.CanUpdate(newDom); err != nil {
			return err
		}

		updated, err := mapper.UpdateReqToRepo(current, req)
		if err != nil {
			return err
		}
		return s.repo.Update(txCtx, updated)
	})
	if err != nil {
		return nil, handler.ToStatus(err).Err()
	}
	return &emptypb.Empty{}, nil
}

// DeleteStudent handles a gRPC request to delete a student by ID.
func (s *StudentServer) DeleteStudent(ctx context.Context, req *proto.DeleteStudentRequest) (*emptypb.Empty, error) {
	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.Delete(txCtx, req.Id); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return d.ErrNotFound
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, handler.ToStatus(err).Err()
	}
	return &emptypb.Empty{}, nil
}

// ListStudents handles a gRPC request to return students by grade.
func (s *StudentServer) ListStudents(ctx context.Context, req *proto.ListStudentsRequest) (*proto.ListStudentsResponse, error) {
	items, err := s.repo.ListByGrade(ctx, req.Grade)
	if err != nil {
		return nil, handler.ToStatus(err).Err()
	}
	out := make([]*proto.Student, 0, len(items))
	for i := range items {
		out = append(out, mapper.RepoToProto(&items[i]))
	}
	return &proto.ListStudentsResponse{Students: out}, nil
}
