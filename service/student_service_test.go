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
	_ "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type fixedIDGenerator struct{}

func (f fixedIDGenerator) GenerateID() string {
	return "generated-id"
}

// StudentServiceTestSuite defines the test suite structure
type StudentServiceTestSuite struct {
	suite.Suite

	ctrl     *gomock.Controller
	mockRepo *mocks.MockRepository
	mockTime *mocks.MockTimeProvider
	mockTx   *mocks.MockTxManager
	server   *service.StudentServer
}

func (s *StudentServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())

	s.mockRepo = mocks.NewMockRepository(s.ctrl)
	s.mockTime = mocks.NewMockTimeProvider(s.ctrl)
	s.mockTx = mocks.NewMockTxManager(s.ctrl)

	idGen := fixedIDGenerator{}
	s.server = service.NewStudentServer(s.mockRepo, s.mockTime, s.mockTx, idGen)
}

func (s *StudentServiceTestSuite) TestSuiteRuns() {
	s.T().Log("Test suite is running")
	s.True(true)
}

func (s *StudentServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *StudentServiceTestSuite) TestCreateStudent_Success() {
	ctx := context.Background()

	req := &proto.CreateStudentRequest{
		FirstName: "Ivan",
		LastName:  "Petrov",
		Grade:     9,
	}
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedStudent := student.Student{
		ID:        "generated-id",
		FirstName: "Ivan",
		LastName:  "Petrov",
		Grade:     9,
		CreatedAt: createdAt,
	}

	s.mockTime.EXPECT().Now().Return(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	s.mockTx.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)
	s.mockRepo.EXPECT().Create(ctx, gomock.Eq(expectedStudent)).Return("generated-id", nil)

	resp, err := s.server.CreateStudent(ctx, req)

	require.NoError(s.T(), err)
	require.Equal(s.T(), "generated-id", resp.GetId())
}

func (s *StudentServiceTestSuite) TestCreateStudent_DBError() {
	ctx := context.Background()
	req := &proto.CreateStudentRequest{FirstName: "Fail", LastName: "Case", Grade: 10}

	s.mockTime.EXPECT().Now().Return(time.Now()).AnyTimes()
	s.mockTx.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})
	s.mockRepo.EXPECT().Create(ctx, gomock.Any()).Return("", errors.New("db error"))

	_, err := s.server.CreateStudent(ctx, req)
	require.Error(s.T(), err)
	st, ok := status.FromError(err)
	require.True(s.T(), ok)
	require.Equal(s.T(), codes.Internal, st.Code())
}

func (s *StudentServiceTestSuite) TestGetStudent_Success() {
	id := "student-id"
	expectedTime := time.Now()

	s.mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(&student.Student{
		ID:        id,
		FirstName: "Alice",
		LastName:  "Smith",
		Grade:     9,
		CreatedAt: expectedTime,
	}, nil)

	resp, err := s.server.GetStudent(context.Background(), &proto.GetStudentRequest{Id: id})
	require.NoError(s.T(), err)
	require.Equal(s.T(), id, resp.Student.Id)
	require.Equal(s.T(), "Alice", resp.Student.FirstName)
	require.Equal(s.T(), "Smith", resp.Student.LastName)
	require.Equal(s.T(), int32(9), resp.Student.Grade)
	require.WithinDuration(s.T(), expectedTime, resp.Student.CreatedAt.AsTime(), time.Second)
}

func (s *StudentServiceTestSuite) TestGetStudent_NotFound() {
	studentID := "non-existent-id"

	s.mockRepo.EXPECT().
		GetByID(gomock.Any(), studentID).
		Return(nil, student.ErrNotFound)

	_, err := s.server.GetStudent(context.Background(), &proto.GetStudentRequest{Id: studentID})

	require.Error(s.T(), err)

	st, ok := status.FromError(err)
	require.True(s.T(), ok)
	require.Equal(s.T(), codes.NotFound, st.Code())
}

func (s *StudentServiceTestSuite) TestGetStudent_DBError() {
	studentID := "some-id"

	s.mockTime.EXPECT().Now().Return(time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)).AnyTimes()
	s.mockRepo.EXPECT().
		GetByID(gomock.Any(), studentID).
		Return(nil, errors.New("db error"))

	_, err := s.server.GetStudent(context.Background(), &proto.GetStudentRequest{Id: studentID})

	require.Error(s.T(), err)

	st, ok := status.FromError(err)
	require.True(s.T(), ok)
	require.Equal(s.T(), codes.Internal, st.Code())
}

func (s *StudentServiceTestSuite) TestUpdateStudent_Success() {
	ctx := context.Background()

	fixedTime := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)
	s.mockTime.EXPECT().Now().Return(fixedTime).AnyTimes()

	s.mockTx.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)

	existingStudent := &student.Student{
		ID:        "some-id",
		FirstName: "Dobby",
		LastName:  "Surname",
		Grade:     9,
		CreatedAt: fixedTime,
	}
	updatedStudent := student.Student{
		ID:        "some-id",
		FirstName: "John",
		LastName:  "Doe",
		Grade:     10,
		CreatedAt: fixedTime,
	}

	s.mockRepo.EXPECT().GetByID(gomock.Any(), "some-id").Return(existingStudent, nil)
	s.mockRepo.EXPECT().Update(gomock.Any(), gomock.Eq(updatedStudent)).Return(nil)

	req := &proto.UpdateStudentRequest{
		Student: &proto.Student{
			Id:        "some-id",
			FirstName: "John",
			LastName:  "Doe",
			Grade:     10,
		},
	}

	_, err := s.server.UpdateStudent(ctx, req)
	require.NoError(s.T(), err)
}

func (s *StudentServiceTestSuite) TestUpdateStudent_DBError() {
	ctx := context.Background()

	fixedTime := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)
	s.mockTime.EXPECT().Now().Return(fixedTime).AnyTimes()

	s.mockTx.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)

	s.mockRepo.EXPECT().
		GetByID(gomock.Any(), "some-id").
		Return(&student.Student{
			ID:        "some-id",
			FirstName: "Dobby",
			LastName:  "Surname",
			Grade:     9,
			CreatedAt: fixedTime,
		}, nil)

	s.mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("db error"))

	req := &proto.UpdateStudentRequest{
		Student: &proto.Student{
			Id:        "some-id",
			FirstName: "John",
			LastName:  "Doe",
			Grade:     10,
		},
	}

	_, err := s.server.UpdateStudent(ctx, req)

	require.Error(s.T(), err)

	st, ok := status.FromError(err)
	require.True(s.T(), ok)
	require.Equal(s.T(), codes.Internal, st.Code())
}

func (s *StudentServiceTestSuite) TestDeleteStudent_Success() {
	ctx := context.Background()
	studentID := "some-id"
	fixedTime := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)

	s.mockTime.EXPECT().Now().Return(fixedTime).AnyTimes()
	s.mockTx.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)
	s.mockRepo.EXPECT().Delete(gomock.Any(), studentID).Return(nil)

	_, err := s.server.DeleteStudent(ctx, &proto.DeleteStudentRequest{Id: studentID})
	require.NoError(s.T(), err)
}

func (s *StudentServiceTestSuite) TestDeleteStudent_DBError() {
	ctx := context.Background()
	studentID := "some-id"
	fixedTime := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)

	s.mockTime.EXPECT().Now().Return(fixedTime).AnyTimes()
	s.mockTx.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)
	s.mockRepo.EXPECT().Delete(gomock.Any(), studentID).Return(errors.New("db error"))

	_, err := s.server.DeleteStudent(ctx, &proto.DeleteStudentRequest{Id: studentID})
	require.Error(s.T(), err)

	st, ok := status.FromError(err)
	require.True(s.T(), ok)
	require.Equal(s.T(), codes.Internal, st.Code())
}

func (s *StudentServiceTestSuite) TestListStudents_Success() {
	ctx := context.Background()
	grade := int32(10)
	students := []student.Student{
		{ID: "1", FirstName: "John", LastName: "Doe", Grade: 10},
		{ID: "2", FirstName: "Alice", LastName: "Smith", Grade: 10},
	}

	s.mockRepo.EXPECT().ListByGrade(gomock.Any(), grade).Return(students, nil)

	resp, err := s.server.ListStudents(ctx, &proto.ListStudentsRequest{Grade: grade})
	require.NoError(s.T(), err)
	require.Len(s.T(), resp.Students, len(students))
}

func (s *StudentServiceTestSuite) TestListStudents_EmptyList() {
	ctx := context.Background()
	grade := int32(10)

	s.mockRepo.EXPECT().ListByGrade(gomock.Any(), grade).Return([]student.Student{}, nil)

	resp, err := s.server.ListStudents(ctx, &proto.ListStudentsRequest{Grade: grade})
	require.NoError(s.T(), err)
	require.Empty(s.T(), resp.Students)
}

func (s *StudentServiceTestSuite) TestListStudents_DBError() {
	ctx := context.Background()
	grade := int32(10)

	s.mockRepo.EXPECT().ListByGrade(gomock.Any(), grade).Return(nil, errors.New("db error"))

	_, err := s.server.ListStudents(ctx, &proto.ListStudentsRequest{Grade: grade})
	require.Error(s.T(), err)

	st, ok := status.FromError(err)
	require.True(s.T(), ok)
	require.Equal(s.T(), codes.Internal, st.Code())
}

func TestStudentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(StudentServiceTestSuite))
}
