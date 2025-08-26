package handler

import (
	"context"

	d "example.com/student-service/domain"
	pb "example.com/student-service/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// StudentServer wires proto requests to the domain service.
type StudentServer struct {
	pb.UnimplementedStudentServiceServer
	svc d.StudentService
}

// NewStudentServer creates a new gRPC handler instance.
func NewStudentServer(svc d.StudentService) *StudentServer {
	return &StudentServer{svc: svc}
}

func (h *StudentServer) CreateStudent(ctx context.Context, req *pb.CreateStudentRequest) (*pb.CreateStudentResponse, error) {
	st := ProtoToDomainStudent(req.GetStudent())
	id, err := h.svc.Create(ctx, st)
	if err != nil {
		return nil, ToStatus(err).Err()
	}
	return &pb.CreateStudentResponse{Id: id}, nil
}

func (h *StudentServer) GetStudent(ctx context.Context, req *pb.GetStudentRequest) (*pb.GetStudentResponse, error) {
	st, err := h.svc.Get(ctx, req.GetId())
	if err != nil {
		return nil, ToStatus(err).Err()
	}
	return &pb.GetStudentResponse{Student: DomainToProtoStudent(*st)}, nil
}

func (h *StudentServer) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*emptypb.Empty, error) {
	st := ProtoToDomainStudent(req.GetStudent())
	if err := h.svc.Update(ctx, st); err != nil {
		return nil, ToStatus(err).Err()
	}
	return &emptypb.Empty{}, nil
}

func (h *StudentServer) DeleteStudent(ctx context.Context, req *pb.DeleteStudentRequest) (*emptypb.Empty, error) {
	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, ToStatus(err).Err()
	}
	return &emptypb.Empty{}, nil
}

func (h *StudentServer) ListStudents(ctx context.Context, req *pb.ListStudentsRequest) (*pb.ListStudentsResponse, error) {
	var (
		list []d.Student
		err  error
	)
	if grade := req.GetGrade(); grade > 0 {
		list, err = h.svc.ListByGrade(ctx, grade)
	} else {
		list, err = h.svc.List(ctx)
	}
	if err != nil {
		return nil, ToStatus(err).Err()
	}

	out := make([]*pb.Student, 0, len(list))
	for _, s := range list {
		out = append(out, DomainToProtoStudent(s))
	}
	return &pb.ListStudentsResponse{Students: out}, nil
}
