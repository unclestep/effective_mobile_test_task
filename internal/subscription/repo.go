package subscription

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("record not found")

type SubscriptionRepo interface {
	Get(ctx context.Context, opt ...SubscriptionGetOpt) ([]*Subscription, error)
	Create(ctx context.Context, subscription *Subscription) error
	Update(ctx context.Context, subscription *Subscription) error
	Delete(ctx context.Context, subscriptionID string) error
}

type SubscriptionGetOpt func(*SubscriptionGetConfig)

type SubscriptionGetConfig struct {
	UserID      *string
	ServiceName *string
	StartPeriod *time.Time
	EndPeriod   *time.Time
}

func WithUserID(userID string) SubscriptionGetOpt {
	return func(cfg *SubscriptionGetConfig) {
		cfg.UserID = &userID
	}
}

func WithServiceName(serviceName string) SubscriptionGetOpt {
	return func(cfg *SubscriptionGetConfig) {
		cfg.ServiceName = &serviceName
	}
}

func WithStartPeriod(start time.Time) SubscriptionGetOpt {
	return func(cfg *SubscriptionGetConfig) {
		cfg.StartPeriod = &start
	}
}

func WithEndPeriod(end time.Time) SubscriptionGetOpt {
	return func(cfg *SubscriptionGetConfig) {
		cfg.EndPeriod = &end
	}
}
