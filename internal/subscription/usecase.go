package subscription

import "context"

type UseCase interface {
	Get(ctx context.Context, opt ...SubscriptionGetOpt) ([]*Subscription, error)
	Create(ctx context.Context, subscription *Subscription) error
	Update(ctx context.Context, subscription *Subscription) error
	Delete(ctx context.Context, subscriptionID string) error
}
