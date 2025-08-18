// Package service implements the student gRPC service logic.
package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"example.com/student-service/proto"
	"example.com/student-service/repository/student"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	studentEntity := student.Student{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Grade:     req.Grade,
		CreatedAt: now,

		MiddleName:   nullStringPtr(req.MiddleName),
		Status:       nullStringFromEnum(req.Status),
		HomeAddress:  nullStringFromAddress(req.HomeAddress),
		CourseGrades: nullStringFromMap(req.CourseGrades),
		Friends:      req.Friends,
	}

	switch details := req.GetStudentDetails().(type) {
	case *proto.CreateStudentRequest_Local:
		if details.Local != nil {
			b, err := json.Marshal(details.Local)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid local data: %v", err)
			}
			studentEntity.Local = sql.NullString{String: string(b), Valid: true}
		}
	case *proto.CreateStudentRequest_Exchange:
		if details.Exchange != nil {
			b, err := json.Marshal(details.Exchange)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid exchange data: %v", err)
			}
			studentEntity.Exchange = sql.NullString{String: string(b), Valid: true}
		}
	}

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		_, err := s.repo.Create(txCtx, studentEntity)
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

	pbStudent := &proto.Student{
		Id:        st.ID,
		FirstName: st.FirstName,
		LastName:  st.LastName,
		Grade:     st.Grade,
		CreatedAt: timestamppb.New(st.CreatedAt),
		Friends:   st.Friends,
	}

	if st.MiddleName.Valid {
		pbStudent.MiddleName = &st.MiddleName.String
	}

	if st.Status.Valid {
		if statusEnum, ok := proto.Student_Status_value[st.Status.String]; ok {
			pbStudent.Status = proto.Student_Status(statusEnum)
		}
	}

	if st.HomeAddress.Valid {
		parts := strings.SplitN(st.HomeAddress.String, ",", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid home address %q: expected 'street, city, country'", st.HomeAddress.String)
		}

		street := strings.TrimSpace(parts[0])
		city := strings.TrimSpace(parts[1])
		country := strings.TrimSpace(parts[2])

		pbStudent.HomeAddress.Street = street
		pbStudent.HomeAddress.City = city
		pbStudent.HomeAddress.Country = country
	}

	if st.CourseGrades.Valid {
		var courseMap map[string]string
		if err := json.Unmarshal([]byte(st.CourseGrades.String), &courseMap); err == nil {
			pbStudent.CourseGrades = courseMap
		}
	}

	if st.Local.Valid {
		var local proto.Student_LocalStudent
		if err := json.Unmarshal([]byte(st.Local.String), &local); err == nil {
			pbStudent.StudentDetails = &proto.Student_Local{Local: &local}
		}
	}

	if st.Exchange.Valid {
		var exchange proto.Student_ExchangeStudent
		if err := json.Unmarshal([]byte(st.Exchange.String), &exchange); err == nil {
			pbStudent.StudentDetails = &proto.Student_Exchange{Exchange: &exchange}
		}
	}

	return &proto.GetStudentResponse{Student: pbStudent}, nil
}

// UpdateStudent handles a gRPC request to update an existing student's data.
func (s *StudentServer) UpdateStudent(ctx context.Context, req *proto.UpdateStudentRequest) (*emptypb.Empty, error) {
	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
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
			ID:           req.Student.Id,
			FirstName:    req.Student.FirstName,
			LastName:     req.Student.LastName,
			Grade:        req.Student.Grade,
			CreatedAt:    existing.CreatedAt,
			MiddleName:   nullStringPtr(req.Student.MiddleName),
			Status:       nullStringFromEnum(req.Student.Status),
			HomeAddress:  nullStringFromAddress(req.Student.HomeAddress),
			CourseGrades: nullStringFromMap(req.Student.CourseGrades),
			Friends:      req.Student.Friends,
			Local:        nullStringFromLocal(req.Student.GetLocal()),
			Exchange:     nullStringFromExchange(req.Student.GetExchange()),
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
	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
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
		pbStudent := &proto.Student{
			Id:           st.ID,
			FirstName:    st.FirstName,
			LastName:     st.LastName,
			Grade:        st.Grade,
			CreatedAt:    timestamppb.New(st.CreatedAt),
			MiddleName:   nullToPtr(st.MiddleName),
			Status:       parseStatus(st.Status),
			HomeAddress:  parseAddress(st.HomeAddress),
			CourseGrades: parseCourseGrades(st.CourseGrades),
			Friends:      st.Friends,
		}

		if st.Local.Valid {
			local := &proto.Student_LocalStudent{}
			if err := json.Unmarshal([]byte(st.Local.String), local); err == nil {
				pbStudent.StudentDetails = &proto.Student_Local{Local: local}
			}
		} else if st.Exchange.Valid {
			exchange := &proto.Student_ExchangeStudent{}
			if err := json.Unmarshal([]byte(st.Exchange.String), exchange); err == nil {
				pbStudent.StudentDetails = &proto.Student_Exchange{Exchange: exchange}
			}
		}

		result = append(result, pbStudent)
	}

	return &proto.ListStudentsResponse{Students: result}, nil
}

func nullStringPtr(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullStringFromEnum(s proto.Student_Status) sql.NullString {
	if s == proto.Student_STATUS_UNSPECIFIED {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s.String(), Valid: true}
}

func nullStringFromAddress(addr *proto.Student_Address) sql.NullString {
	if addr == nil {
		return sql.NullString{Valid: false}
	}
	full := fmt.Sprintf("%s, %s, %s", addr.Street, addr.City, addr.Country)
	return sql.NullString{String: full, Valid: full != ""}
}

func nullStringFromMap(m map[string]string) sql.NullString {
	if len(m) == 0 {
		return sql.NullString{Valid: false}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: string(b), Valid: true}
}

func nullStringFromLocal(local *proto.Student_LocalStudent) sql.NullString {
	if local == nil {
		return sql.NullString{Valid: false}
	}
	bytes, err := json.Marshal(local)
	if err != nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: string(bytes), Valid: true}
}

func nullStringFromExchange(exchange *proto.Student_ExchangeStudent) sql.NullString {
	if exchange == nil {
		return sql.NullString{Valid: false}
	}
	bytes, err := json.Marshal(exchange)
	if err != nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: string(bytes), Valid: true}
}

func nullToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func parseStatus(ns sql.NullString) proto.Student_Status {
	if !ns.Valid {
		return proto.Student_STATUS_UNSPECIFIED
	}
	status, ok := proto.Student_Status_value[ns.String]
	if !ok {
		return proto.Student_STATUS_UNSPECIFIED
	}
	return proto.Student_Status(status)
}

func parseAddress(ns sql.NullString) *proto.Student_Address {
	if !ns.Valid {
		return nil
	}

	parts := strings.SplitN(ns.String, ",", 3)
	if len(parts) != 3 {
		return &proto.Student_Address{}
	}

	return &proto.Student_Address{
		Street:  strings.TrimSpace(parts[0]),
		City:    strings.TrimSpace(parts[1]),
		Country: strings.TrimSpace(parts[2]),
	}
}

func parseCourseGrades(ns sql.NullString) map[string]string {
	if !ns.Valid {
		return nil
	}
	var grades map[string]string
	_ = json.Unmarshal([]byte(ns.String), &grades)
	return grades
}
