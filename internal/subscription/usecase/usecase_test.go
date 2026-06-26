package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"em/internal/subscription"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Get(ctx context.Context, filter *subscription.SubscriptionFilter) ([]*subscription.Subscription, error) {
	args := m.Called(ctx, filter)
	var subs []*subscription.Subscription
	if v := args.Get(0); v != nil {
		subs = v.([]*subscription.Subscription)
	}
	return subs, args.Error(1)
}

func (m *MockRepo) GetTotalCost(ctx context.Context, filter *subscription.SubscriptionFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}

func (m *MockRepo) Create(ctx context.Context, sub *subscription.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *MockRepo) Update(ctx context.Context, sub *subscription.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *MockRepo) Delete(ctx context.Context, subID string) error {
	return m.Called(ctx, subID).Error(0)
}

func newTestUC(repo subscription.SubscriptionRepo) subscription.UseCase {
	return NewSubscriptionUC(zap.NewNop(), repo)
}

func TestGetSuccess(t *testing.T) {
	repo := new(MockRepo)
	filter := &subscription.SubscriptionFilter{}
	want := []*subscription.Subscription{
		{SubscriptionID: "id-1", ServiceName: "Yandex Plus", Price: 400},
	}
	repo.On("Get", mock.Anything, filter).Return(want, nil)

	uc := newTestUC(repo)
	got, err := uc.Get(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	repo.AssertExpectations(t)
}

func TestGetEmptyIsNotAnError(t *testing.T) {
	repo := new(MockRepo)
	filter := &subscription.SubscriptionFilter{UserID: new("user-without-subs")}
	repo.On("Get", mock.Anything, filter).Return(nil, nil)

	uc := newTestUC(repo)
	got, err := uc.Get(context.Background(), filter)

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetRepoError(t *testing.T) {
	repo := new(MockRepo)
	repoErr := errors.New("connection refused")
	filter := &subscription.SubscriptionFilter{}
	repo.On("Get", mock.Anything, filter).Return(nil, repoErr)

	uc := newTestUC(repo)
	got, err := uc.Get(context.Background(), filter)

	assert.Nil(t, got)
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}

func TestGetByIDFound(t *testing.T) {
	repo := new(MockRepo)
	subID := "id-1"
	want := &subscription.Subscription{SubscriptionID: subID, ServiceName: "Yandex Plus"}
	repo.On("Get", mock.Anything, &subscription.SubscriptionFilter{SubID: &subID}).
		Return([]*subscription.Subscription{want}, nil)

	uc := newTestUC(repo)
	got, err := uc.GetByID(context.Background(), subID)

	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestGetByIDNotFound(t *testing.T) {
	repo := new(MockRepo)
	subID := "missing-id"
	repo.On("Get", mock.Anything, &subscription.SubscriptionFilter{SubID: &subID}).
		Return(nil, nil)

	uc := newTestUC(repo)
	got, err := uc.GetByID(context.Background(), subID)

	assert.Nil(t, got)
	assert.ErrorIs(t, err, subscription.ErrNotFound)
}

func TestGetByIDRepoError(t *testing.T) {
	repo := new(MockRepo)
	subID := "id-1"
	repoErr := errors.New("db unavailable")
	repo.On("Get", mock.Anything, &subscription.SubscriptionFilter{SubID: &subID}).
		Return(nil, repoErr)

	uc := newTestUC(repo)
	got, err := uc.GetByID(context.Background(), subID)

	assert.Nil(t, got)
	assert.ErrorIs(t, err, repoErr)
}

func TestGetTotalCostSuccess(t *testing.T) {
	repo := new(MockRepo)
	filter := &subscription.SubscriptionFilter{}
	repo.On("GetTotalCost", mock.Anything, filter).Return(1200, nil)

	uc := newTestUC(repo)
	total, err := uc.GetTotalCost(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, 1200, total)
}

func TestGetTotalCostEmpty(t *testing.T) {
	repo := new(MockRepo)
	filter := &subscription.SubscriptionFilter{UserID: new("user-without-subs")}
	repo.On("GetTotalCost", mock.Anything, filter).Return(0, nil)

	uc := newTestUC(repo)
	total, err := uc.GetTotalCost(context.Background(), filter)

	require.NoError(t, err)
	assert.Zero(t, total)
}

func TestGetTotalCostRepoError(t *testing.T) {
	repo := new(MockRepo)
	repoErr := errors.New("query failed")
	filter := &subscription.SubscriptionFilter{}
	repo.On("GetTotalCost", mock.Anything, filter).Return(0, repoErr)

	uc := newTestUC(repo)
	total, err := uc.GetTotalCost(context.Background(), filter)

	assert.Zero(t, total)
	assert.ErrorIs(t, err, repoErr)
}

func TestCreateSuccess(t *testing.T) {
	repo := new(MockRepo)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*subscription.Subscription")).Return(nil)

	uc := newTestUC(repo)
	input := &subscription.CreateSubInput{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "user-1",
		StartDate:   time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	got, err := uc.Create(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.NotEmpty(t, got.SubscriptionID)
	_, uuidErr := uuid.Parse(got.SubscriptionID)
	assert.NoError(t, uuidErr)
	assert.Equal(t, input.ServiceName, got.ServiceName)
	assert.Equal(t, input.Price, got.Price)
	assert.Nil(t, got.EndDate)
	repo.AssertExpectations(t)
}

func TestCreateRepoError(t *testing.T) {
	repo := new(MockRepo)
	repoErr := errors.New("constraint violation")
	repo.On("Create", mock.Anything, mock.AnythingOfType("*subscription.Subscription")).Return(repoErr)

	uc := newTestUC(repo)
	input := &subscription.CreateSubInput{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "user-1",
		StartDate:   time.Now(),
	}

	got, err := uc.Create(context.Background(), input)

	assert.Nil(t, got)
	assert.ErrorIs(t, err, repoErr)
}

func TestUpdateNilInput(t *testing.T) {
	repo := new(MockRepo)
	uc := newTestUC(repo)

	err := uc.Update(context.Background(), "id-1", nil)

	assert.ErrorIs(t, err, subscription.ErrNothingToUpdate)
	repo.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
}

func TestUpdateAllFieldsNil(t *testing.T) {
	repo := new(MockRepo)
	uc := newTestUC(repo)

	err := uc.Update(context.Background(), "id-1", &subscription.UpdateSubInput{})

	assert.ErrorIs(t, err, subscription.ErrNothingToUpdate)
	repo.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
}

func TestUpdateNotFound(t *testing.T) {
	repo := new(MockRepo)
	subID := "missing-id"
	repo.On("Get", mock.Anything, &subscription.SubscriptionFilter{SubID: &subID}).
		Return(nil, nil)

	uc := newTestUC(repo)
	err := uc.Update(context.Background(), subID, &subscription.UpdateSubInput{Price: new(500)})

	assert.ErrorIs(t, err, subscription.ErrNotFound)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdateGetRepoError(t *testing.T) {
	repo := new(MockRepo)
	subID := "id-1"
	repoErr := errors.New("connection lost")
	repo.On("Get", mock.Anything, &subscription.SubscriptionFilter{SubID: &subID}).
		Return(nil, repoErr)

	uc := newTestUC(repo)
	err := uc.Update(context.Background(), subID, &subscription.UpdateSubInput{Price: new(500)})

	assert.ErrorIs(t, err, repoErr)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdateSuccessPartialPatch(t *testing.T) {
	repo := new(MockRepo)
	subID := "id-1"
	existing := &subscription.Subscription{
		SubscriptionID: subID,
		ServiceName:    "Yandex Plus",
		Price:          400,
		UserID:         "user-1",
		StartDate:      time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
		EndDate:        nil,
	}
	repo.On("Get", mock.Anything, &subscription.SubscriptionFilter{SubID: &subID}).
		Return([]*subscription.Subscription{existing}, nil)

	newPrice := 600
	repo.On("Update", mock.Anything, mock.MatchedBy(func(s *subscription.Subscription) bool {
		return s.SubscriptionID == subID &&
			s.Price == newPrice &&
			s.ServiceName == "Yandex Plus" &&
			s.UserID == "user-1" &&
			s.StartDate.Equal(existing.StartDate) &&
			s.EndDate == nil
	})).Return(nil)

	uc := newTestUC(repo)
	err := uc.Update(context.Background(), subID, &subscription.UpdateSubInput{Price: &newPrice})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateUpdateRepoError(t *testing.T) {
	repo := new(MockRepo)
	subID := "id-1"
	existing := &subscription.Subscription{SubscriptionID: subID, Price: 400}
	repo.On("Get", mock.Anything, &subscription.SubscriptionFilter{SubID: &subID}).
		Return([]*subscription.Subscription{existing}, nil)

	repoErr := errors.New("constraint violation")
	repo.On("Update", mock.Anything, mock.AnythingOfType("*subscription.Subscription")).Return(repoErr)

	uc := newTestUC(repo)
	err := uc.Update(context.Background(), subID, &subscription.UpdateSubInput{Price: new(600)})

	assert.ErrorIs(t, err, repoErr)
}

func TestDeleteSuccess(t *testing.T) {
	repo := new(MockRepo)
	repo.On("Delete", mock.Anything, "id-1").Return(nil)

	uc := newTestUC(repo)
	err := uc.Delete(context.Background(), "id-1")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteNotFound(t *testing.T) {
	repo := new(MockRepo)
	repo.On("Delete", mock.Anything, "missing-id").Return(subscription.ErrNotFound)

	uc := newTestUC(repo)
	err := uc.Delete(context.Background(), "missing-id")

	assert.ErrorIs(t, err, subscription.ErrNotFound)
}
