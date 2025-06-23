package server

import (
    "context"
    "errors"
    "sync"
    "time"

    "github.com/google/uuid"
    "example.com/student-service/proto" 
    "google.golang.org/protobuf/types/known/emptypb"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type Student struct {
    ID        string
    FirstName string
    LastName  string
    Grade     int32
    CreatedAt time.Time
}

type Store struct {
    mu       sync.RWMutex
    students map[string]Student
}

type StudentServer struct {
    proto.UnimplementedStudentServiceServer
    store *Store
}

func NewStudentServer() *StudentServer {
    return &StudentServer{
        store: &Store{
            students: make(map[string]Student),
        },
    }
}

func (s *StudentServer) CreateStudent(ctx context.Context, req *proto.CreateStudentRequest) (*proto.CreateStudentResponse, error) {
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

func (s *StudentServer) GetStudent(ctx context.Context, req *proto.GetStudentRequest) (*proto.GetStudentResponse, error) {
    s.store.mu.RLock()
    student, ok := s.store.students[req.Id]
    s.store.mu.RUnlock()

    if !ok {
        return nil, errors.New("student not found")
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

func (s *StudentServer) UpdateStudent(ctx context.Context, req *proto.UpdateStudentRequest) (*emptypb.Empty, error) {
    student := req.Student

    s.store.mu.Lock()
    stored, ok := s.store.students[student.Id]
    if !ok {
        s.store.mu.Unlock()
        return nil, errors.New("student not found")
    }

    if student.Grade < stored.Grade {
        s.store.mu.Unlock()
        return nil, errors.New("grade cannot be decreased")
    }

    stored.FirstName = student.FirstName
    stored.LastName = student.LastName
    stored.Grade = student.Grade

    s.store.students[student.Id] = stored
    s.store.mu.Unlock()

    return &emptypb.Empty{}, nil
}

func (s *StudentServer) DeleteStudent(ctx context.Context, req *proto.DeleteStudentRequest) (*emptypb.Empty, error) {
    s.store.mu.Lock()
    defer s.store.mu.Unlock()

    pupul := req.Id

    if _, ok := s.store.students[pupul]; !ok {
        return nil, errors.New("student not found")
    }

    delete(s.store.students, pupul)
    return &emptypb.Empty{}, nil
}

func (s *StudentServer) ListStudents(ctx context.Context, req *proto.ListStudentsRequest) (*proto.ListStudentsResponse, error) {
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
