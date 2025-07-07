package service_test

import (
	"context"
	"testing"
	"time"

	"example.com/student-service/proto"
	"example.com/student-service/service"
	"example.com/student-service/service/mocks"

	"github.com/golang/mock/gomock"
)

type fixedIDGenerator struct{}

func (f fixedIDGenerator) GenerateID() string {
	return "generated-id"
}

func TestStudentServer_CreateStudent(t *testing.T) {
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
