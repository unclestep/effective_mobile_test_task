package subscription

import (
	"context"
	"errors"
	"time"
)

var ErrNothingToUpdate = errors.New("nothing to update")

type UseCase interface {
	Get(ctx context.Context, filter *SubscriptionFilter) ([]*Subscription, error)
	GetTotalCost(ctx context.Context, filter *SubscriptionFilter) (int, error)
	Create(ctx context.Context, sub *CreateSubInput) error
	Update(ctx context.Context, subID string, sub *UpdateSubInput) error
	Delete(ctx context.Context, subID string) error
}

type SubscriptionFilter struct {
	SubID       *string
	UserID      *string
	ServiceName *string
	StartDate   *time.Time
	EndDate     *time.Time
}

type CreateSubInput struct {
	ServiceName string
	Price       int
	UserID      string
	StartDate   time.Time
	EndDate     *time.Time
}

type UpdateSubInput struct {
	ServiceName *string
	Price       *int
	StartDate   *time.Time
	EndDate     *time.Time
}
