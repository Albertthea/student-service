package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/student-service/proto"
	"example.com/student-service/repository/student"
	"example.com/student-service/service"
	"example.com/student-service/service/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/mock/gomock"
)

type fixedIDGenerator struct{}

func (f fixedIDGenerator) GenerateID() string {
	return "generated-id"
}

func TestStudentServer_CreateStudent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockTime := mocks.NewMockTimeProvider(ctrl)
	mockTx := mocks.NewMockTxManager(ctrl)

	fixedTime := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)
	mockTime.EXPECT().Now().Return(fixedTime).AnyTimes()

	// Expect WithTransaction to invoke the provided function
	mockTx.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)

	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("generated-id", nil)

	idGen := fixedIDGenerator{}
	server := service.NewStudentServer(mockRepo, mockTime, mockTx, idGen)

	req := &proto.CreateStudentRequest{
		FirstName: "John",
		LastName:  "Doe",
		Grade:     10,
	}

	resp, err := server.CreateStudent(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Id != "generated-id" {
		t.Errorf("expected id 'generated-id', got %s", resp.Id)
	}
}

func TestStudentServer_CreateStudent_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockTime := mocks.NewMockTimeProvider(ctrl)
	mockTx := mocks.NewMockTxManager(ctrl)

	fixedTime := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)
	mockTime.EXPECT().Now().Return(fixedTime).AnyTimes()

	mockTx.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)

	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("", errors.New("db error"))

	idGen := fixedIDGenerator{}
	server := service.NewStudentServer(mockRepo, mockTime, mockTx, idGen)

	req := &proto.CreateStudentRequest{
		FirstName: "Fail",
		LastName:  "Case",
		Grade:     10,
	}

	_, err := server.CreateStudent(context.Background(), req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Internal {
		t.Fatalf("gRPC Internal error expected, but received: %v", err)
	}
}

func TestStudentServer_GetStudent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockTime := mocks.NewMockTimeProvider(ctrl)
	mockTx := mocks.NewMockTxManager(ctrl)

	idGen := fixedIDGenerator{}
	server := service.NewStudentServer(mockRepo, mockTime, mockTx, idGen)

	expectedTime := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)
	studentID := "some-id"

	mockRepo.EXPECT().GetByID(gomock.Any(), studentID).Return(&student.Student{
		ID:        studentID,
		FirstName: "Alice",
		LastName:  "Smith",
		Grade:     9,
		CreatedAt: expectedTime,
	}, nil)

	resp, err := server.GetStudent(context.Background(), &proto.GetStudentRequest{Id: studentID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Student == nil {
		t.Fatal("expected student, got nil")
	}

	if resp.Student.Id != studentID {
		t.Errorf("expected id %s, got %s", studentID, resp.Student.Id)
	}
	if resp.Student.FirstName != "Alice" {
		t.Errorf("expected first name Alice, got %s", resp.Student.FirstName)
	}
	if resp.Student.LastName != "Smith" {
		t.Errorf("expected last name Smith, got %s", resp.Student.LastName)
	}
	if resp.Student.Grade != 9 {
		t.Errorf("expected grade 9, got %d", resp.Student.Grade)
	}
	if !resp.Student.CreatedAt.AsTime().Equal(expectedTime) {
		t.Errorf("expected CreatedAt %v, got %v", expectedTime, resp.Student.CreatedAt.AsTime())
	}
}

func TestStudentServer_GetStudent_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockTime := mocks.NewMockTimeProvider(ctrl)
	mockTx := mocks.NewMockTxManager(ctrl)

	idGen := fixedIDGenerator{}
	server := service.NewStudentServer(mockRepo, mockTime, mockTx, idGen)

	studentID := "non-existent-id"

	mockRepo.EXPECT().
		GetByID(gomock.Any(), studentID).
		Return(nil, student.ErrNotFound)

	_, err := server.GetStudent(context.Background(), &proto.GetStudentRequest{Id: studentID})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound error, got %v", err)
	}
}

func TestStudentServer_GetStudent_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockTime := mocks.NewMockTimeProvider(ctrl)
	mockTx := mocks.NewMockTxManager(ctrl)

	studentID := "some-id"
	idGen := fixedIDGenerator{}
	server := service.NewStudentServer(mockRepo, mockTime, mockTx, idGen)

	fixedTime := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)
	mockTime.EXPECT().Now().Return(fixedTime).AnyTimes()

	mockRepo.EXPECT().
		GetByID(gomock.Any(), studentID).
		Return(nil, errors.New("db error"))

	_, err := server.GetStudent(context.Background(), &proto.GetStudentRequest{Id: studentID})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Internal {
		t.Fatalf("gRPC Internal error expected, but received: %v", err)
	}
}

func TestStudentServer_UpdateStudent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockTime := mocks.NewMockTimeProvider(ctrl)
	mockTx := mocks.NewMockTxManager(ctrl)

	fixedTime := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)
	mockTime.EXPECT().Now().Return(fixedTime).AnyTimes()

	mockTx.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)

	mockRepo.EXPECT().
		GetByID(gomock.Any(), "some-id").
		Return(&student.Student{
			ID:        "some-id",
			FirstName: "Dobby",
			LastName:  "Surname",
			Grade:     9,
			CreatedAt: fixedTime,
		}, nil)

	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	idGen := fixedIDGenerator{}
	server := service.NewStudentServer(mockRepo, mockTime, mockTx, idGen)

	req := &proto.UpdateStudentRequest{
		Student: &proto.Student{
			Id:        "some-id",
			FirstName: "John",
			LastName:  "Doe",
			Grade:     10,
		},
	}

	_, err := server.UpdateStudent(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

}

func TestStudentServer_UpdateStudent_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockTime := mocks.NewMockTimeProvider(ctrl)
	mockTx := mocks.NewMockTxManager(ctrl)

	fixedTime := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)
	mockTime.EXPECT().Now().Return(fixedTime).AnyTimes()

	mockTx.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)

	mockRepo.EXPECT().
		GetByID(gomock.Any(), "some-id").
		Return(&student.Student{
			ID:        "some-id",
			FirstName: "Dobby",
			LastName:  "Surname",
			Grade:     9,
			CreatedAt: fixedTime,
		}, nil)

	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("db error"))

	idGen := fixedIDGenerator{}
	server := service.NewStudentServer(mockRepo, mockTime, mockTx, idGen)

	req := &proto.UpdateStudentRequest{
		Student: &proto.Student{
			Id:        "some-id",
			FirstName: "John",
			LastName:  "Doe",
			Grade:     10,
		},
	}

	_, err := server.UpdateStudent(context.Background(), req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Internal {
		t.Fatalf("gRPC Internal error expected, but received: %v", err)
	}
}
