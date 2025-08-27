// Package service_test contains tests for the student service business logic.
package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	d "example.com/student-service/domain"
	"example.com/student-service/service"
	"example.com/student-service/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type fixedIDGenerator struct{}

func (f fixedIDGenerator) GenerateID() string { return "generated-id" }

type StudentServiceTestSuite struct {
	suite.Suite

	ctrl     *gomock.Controller
	mockRepo *mocks.MockRepository
	mockTime *mocks.MockTimeProvider
	mockTx   *mocks.MockTxManager
	svc      *service.Service
}

func (s *StudentServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mocks.NewMockRepository(s.ctrl)
	s.mockTime = mocks.NewMockTimeProvider(s.ctrl)
	s.mockTx = mocks.NewMockTxManager(s.ctrl)
	s.svc = service.New(s.mockRepo, s.mockTx, s.mockTime, fixedIDGenerator{})
}

func (s *StudentServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *StudentServiceTestSuite) expectTx() {
	s.mockTx.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
}

func (s *StudentServiceTestSuite) TestSanity() {
	require.NotNil(s.T(), s.svc)
}

// --- Create ---

func (s *StudentServiceTestSuite) TestCreate_Success() {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	in := d.Student{
		FirstName: "Ivan",
		LastName:  "Petrov",
		Grade:     9,
		Status:    d.StatusActive, // важно: валидный статус
		Details:   d.Details{Local: &d.LocalStudent{NationalID: "NID", Scholarship: true}},
	}

	s.mockTime.EXPECT().Now().Return(now)
	s.expectTx()
	s.mockRepo.EXPECT().
		Create(gomock.Any(), gomock.AssignableToTypeOf(d.Student{})).
		DoAndReturn(func(_ context.Context, got d.Student) (string, error) {
			require.Equal(s.T(), "generated-id", got.ID)
			require.Equal(s.T(), "Ivan", got.FirstName)
			require.Equal(s.T(), "Petrov", got.LastName)
			require.EqualValues(s.T(), 9, got.Grade)
			require.Equal(s.T(), d.StatusActive, got.Status)
			require.WithinDuration(s.T(), now, got.CreatedAt, time.Microsecond)
			require.NotNil(s.T(), got.Details.Local)
			require.Equal(s.T(), "NID", got.Details.Local.NationalID)
			return "generated-id", nil
		})

	id, err := s.svc.Create(context.Background(), in)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "generated-id", id)
}

func (s *StudentServiceTestSuite) TestCreate_DBError() {
	now := time.Now().UTC()

	in := d.Student{
		FirstName: "Fail",
		LastName:  "Case",
		Grade:     10,
		Status:    d.StatusActive, // важно
		Details:   d.Details{Local: &d.LocalStudent{NationalID: "X"}},
	}

	s.mockTime.EXPECT().Now().Return(now)
	s.expectTx()
	s.mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return("", errors.New("db error"))

	_, err := s.svc.Create(context.Background(), in)
	require.Error(s.T(), err)
}

// --- Get ---

func (s *StudentServiceTestSuite) TestGet_Success() {
	id := "student-id"
	created := time.Now().UTC()

	s.mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(&d.Student{
		ID:        id,
		FirstName: "Alice",
		LastName:  "Smith",
		Grade:     9,
		Status:    d.StatusActive,
		CreatedAt: created,
	}, nil)

	got, err := s.svc.Get(context.Background(), id)
	require.NoError(s.T(), err)
	require.Equal(s.T(), id, got.ID)
	require.Equal(s.T(), "Alice", got.FirstName)
	require.Equal(s.T(), "Smith", got.LastName)
	require.EqualValues(s.T(), 9, got.Grade)
	require.Equal(s.T(), d.StatusActive, got.Status)
	require.WithinDuration(s.T(), created, got.CreatedAt, time.Second)
}

func (s *StudentServiceTestSuite) TestGet_NotFound() {
	id := "nope"

	s.mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, d.ErrNotFound)
	_, err := s.svc.Get(context.Background(), id)
	require.ErrorIs(s.T(), err, d.ErrNotFound)
}

func (s *StudentServiceTestSuite) TestGet_DBError() {
	id := "oops"

	s.mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, errors.New("db error"))
	_, err := s.svc.Get(context.Background(), id)
	require.Error(s.T(), err)
}

// --- Update ---

func (s *StudentServiceTestSuite) TestUpdate_Success() {
	t0 := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)

	current := &d.Student{
		ID:        "some-id",
		FirstName: "Dobby",
		LastName:  "Surname",
		Grade:     9,
		Status:    d.StatusActive, // важно
		CreatedAt: t0,
		Details:   d.Details{Local: &d.LocalStudent{NationalID: "NID"}},
	}
	upd := d.Student{
		ID:      "some-id",
		Grade:   10,
		Status:  d.StatusActive, // важно
		Details: d.Details{Local: &d.LocalStudent{NationalID: "NID"}},
	}

	s.expectTx()
	s.mockRepo.EXPECT().GetByID(gomock.Any(), "some-id").Return(current, nil)
	s.mockRepo.EXPECT().
		Update(gomock.Any(), gomock.AssignableToTypeOf(d.Student{})).
		DoAndReturn(func(_ context.Context, got d.Student) error {
			require.Equal(s.T(), "some-id", got.ID)
			require.EqualValues(s.T(), 10, got.Grade)
			require.Equal(s.T(), d.StatusActive, got.Status)
			// сервис должен подставить CreatedAt из current
			require.WithinDuration(s.T(), t0, got.CreatedAt, time.Second)
			// details должны сохраниться/быть валидными
			require.NotNil(s.T(), got.Details.Local)
			require.Equal(s.T(), "NID", got.Details.Local.NationalID)
			return nil
		})

	err := s.svc.Update(context.Background(), upd)
	require.NoError(s.T(), err)
}

func (s *StudentServiceTestSuite) TestUpdate_GradeDecrease() {
	t0 := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)

	current := &d.Student{
		ID:        "some-id",
		Grade:     10,
		Status:    d.StatusActive,
		CreatedAt: t0,
		Details:   d.Details{Local: &d.LocalStudent{NationalID: "NID"}},
	}
	upd := d.Student{
		ID:      "some-id",
		Grade:   9, // downgrade → ошибка
		Status:  d.StatusActive,
		Details: d.Details{Local: &d.LocalStudent{NationalID: "NID"}},
	}

	s.expectTx()
	s.mockRepo.EXPECT().GetByID(gomock.Any(), "some-id").Return(current, nil)

	err := s.svc.Update(context.Background(), upd)
	require.ErrorIs(s.T(), err, d.ErrGradeDecrease)
}

func (s *StudentServiceTestSuite) TestUpdate_DBError() {
	t0 := time.Date(2025, 7, 7, 12, 0, 0, 0, time.UTC)

	current := &d.Student{
		ID:        "some-id",
		Grade:     9,
		Status:    d.StatusActive,
		CreatedAt: t0,
		Details:   d.Details{Local: &d.LocalStudent{NationalID: "NID"}},
	}
	upd := d.Student{
		ID:      "some-id",
		Grade:   10,
		Status:  d.StatusActive,
		Details: d.Details{Local: &d.LocalStudent{NationalID: "NID"}},
	}

	s.expectTx()
	s.mockRepo.EXPECT().GetByID(gomock.Any(), "some-id").Return(current, nil)
	s.mockRepo.EXPECT().
		Update(gomock.Any(), gomock.AssignableToTypeOf(d.Student{})).
		Return(errors.New("db error"))

	err := s.svc.Update(context.Background(), upd)
	require.Error(s.T(), err)
}

// --- Delete ---

func (s *StudentServiceTestSuite) TestDelete_Success() {
	id := "some-id"

	s.expectTx()
	s.mockRepo.EXPECT().Delete(gomock.Any(), id).Return(nil)

	err := s.svc.Delete(context.Background(), id)
	require.NoError(s.T(), err)
}

func (s *StudentServiceTestSuite) TestDelete_DBError() {
	id := "some-id"

	s.expectTx()
	s.mockRepo.EXPECT().Delete(gomock.Any(), id).Return(errors.New("db error"))

	err := s.svc.Delete(context.Background(), id)
	require.Error(s.T(), err)
}

// --- List / ListByGrade ---

func (s *StudentServiceTestSuite) TestListByGrade_Success() {
	grade := int32(10)
	items := []d.Student{
		{ID: "1", FirstName: "John", LastName: "Doe", Grade: 10, Status: d.StatusActive},
		{ID: "2", FirstName: "Alice", LastName: "Smith", Grade: 10, Status: d.StatusActive},
	}

	s.mockRepo.EXPECT().ListByGrade(gomock.Any(), grade).Return(items, nil)
	got, err := s.svc.ListByGrade(context.Background(), grade)
	require.NoError(s.T(), err)
	require.Len(s.T(), got, len(items))
}

func (s *StudentServiceTestSuite) TestListByGrade_Empty() {
	grade := int32(10)

	s.mockRepo.EXPECT().ListByGrade(gomock.Any(), grade).Return([]d.Student{}, nil)
	got, err := s.svc.ListByGrade(context.Background(), grade)
	require.NoError(s.T(), err)
	require.Empty(s.T(), got)
}

func (s *StudentServiceTestSuite) TestListByGrade_DBError() {
	grade := int32(10)

	s.mockRepo.EXPECT().ListByGrade(gomock.Any(), grade).Return(nil, errors.New("db error"))
	_, err := s.svc.ListByGrade(context.Background(), grade)
	require.Error(s.T(), err)
}

func (s *StudentServiceTestSuite) TestList_All() {
	items := []d.Student{{ID: "1", FirstName: "Ann", LastName: "Lee", Grade: 7, Status: d.StatusActive}}

	s.mockRepo.EXPECT().List(gomock.Any()).Return(items, nil)
	got, err := s.svc.List(context.Background())
	require.NoError(s.T(), err)
	require.Len(s.T(), got, 1)
}

func TestStudentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(StudentServiceTestSuite))
}
