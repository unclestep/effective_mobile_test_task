package subscription

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("record not found")

type SubscriptionRepo interface {
	Get(ctx context.Context, filter *SubscriptionFilter) ([]*Subscription, error)
	Create(ctx context.Context, subscription *Subscription) error
	Update(ctx context.Context, subscription *Subscription) error
	Delete(ctx context.Context, subscriptionID string) error
}
