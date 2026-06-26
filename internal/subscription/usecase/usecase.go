package usecase

import (
	"context"
	"fmt"

	"em/internal/subscription"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	getMethod          = "get"
	getByIDMethod      = "get by id"
	getTotalCostMethod = "get total cost"
	createMethod       = "create"
	updateMethod       = "update"
	deleteMethod       = "delete"
)

type SubscriptionUC struct {
	logger *zap.Logger
	repo   subscription.SubscriptionRepo
}

func NewSubscriptionUC(logger *zap.Logger, repo subscription.SubscriptionRepo) subscription.UseCase {
	return &SubscriptionUC{
		logger: logger,
		repo:   repo,
	}
}

func (u *SubscriptionUC) Get(ctx context.Context, filter *subscription.SubscriptionFilter) ([]*subscription.Subscription, error) {
	subs, err := u.repo.Get(ctx, filter)
	if err != nil {
		u.logger.Error("failed to get subscriptions", zap.Any("get params", filter), zap.Error(err))
		return nil, wrap(getMethod, err)
	}
	return subs, nil
}

func (u *SubscriptionUC) GetByID(ctx context.Context, subID string) (*subscription.Subscription, error) {
	subs, err := u.repo.Get(ctx, &subscription.SubscriptionFilter{SubID: &subID})
	if err != nil {
		u.logger.Error("failed to get subscription by id", zap.String("subscription_id", subID), zap.Error(err))
		return nil, wrap(getByIDMethod, err)
	}
	if len(subs) == 0 {
		u.logger.Warn("subscription not found", zap.String("subscription_id", subID))
		return nil, wrap(getByIDMethod, subscription.ErrNotFound)
	}
	return subs[0], nil
}

func (u *SubscriptionUC) GetTotalCost(ctx context.Context, filter *subscription.SubscriptionFilter) (int, error) {
	total, err := u.repo.GetTotalCost(ctx, filter)
	if err != nil {
		u.logger.Error("failed to get total cost of subscriptions", zap.Any("get params", filter), zap.Error(err))
		return 0, wrap(getTotalCostMethod, err)
	}
	return total, nil
}

func (u *SubscriptionUC) Create(ctx context.Context, subInp *subscription.CreateSubInput) (*subscription.Subscription, error) {
	sub := subscription.NewSubscription(
		uuid.NewString(), subInp.ServiceName, subInp.Price, subInp.UserID, subInp.StartDate, subInp.EndDate,
	)
	if err := u.repo.Create(ctx, sub); err != nil {
		u.logger.Error("failed to create a new subscription", zap.Any("subscription_input", subInp), zap.Error(err))
		return nil, wrap(createMethod, err)
	}

	u.logger.Info("subscription created", zap.Any("subscription_data", sub))
	return sub, nil
}

func (u *SubscriptionUC) Update(ctx context.Context, subID string, subInp *subscription.UpdateSubInput) error {
	if subInp == nil || (subInp.ServiceName == nil && subInp.Price == nil && subInp.StartDate == nil && subInp.EndDate == nil) {
		u.logger.Warn("nothing was passed for update", zap.Any("subscription_input", subInp))
		return wrap(updateMethod, subscription.ErrNothingToUpdate)
	}

	repoSubs, err := u.repo.Get(ctx, &subscription.SubscriptionFilter{SubID: &subID})
	if err != nil {
		u.logger.Error("failed to fetch subscription for update", zap.String("subscription_id", subID), zap.Error(err))
		return wrap(updateMethod, err)
	}
	if len(repoSubs) == 0 {
		u.logger.Warn("can't find the subscription", zap.String("subscription_id", subID))
		return wrap(updateMethod, subscription.ErrNotFound)
	}
	repoSub := repoSubs[0]

	if subInp.ServiceName != nil {
		repoSub.ServiceName = *subInp.ServiceName
	}
	if subInp.Price != nil {
		repoSub.Price = *subInp.Price
	}
	if subInp.StartDate != nil {
		repoSub.StartDate = *subInp.StartDate
	}
	if subInp.EndDate != nil {
		repoSub.EndDate = subInp.EndDate
	}

	if err := u.repo.Update(ctx, repoSub); err != nil {
		u.logger.Error("failed to update the subscription data", zap.Any("subscription", repoSub), zap.Error(err))
		return wrap(updateMethod, err)
	}

	u.logger.Info("subscription updated", zap.String("subscription_id", subID), zap.Any("subscription_data", repoSub))
	return nil
}

func (u *SubscriptionUC) Delete(ctx context.Context, subID string) error {
	if err := u.repo.Delete(ctx, subID); err != nil {
		u.logger.Error("failed to delete the subscription", zap.String("subscription_id", subID), zap.Error(err))
		return wrap(deleteMethod, err)
	}
	u.logger.Info("subscription deleted", zap.String("subscription_id", subID))
	return nil
}

func wrap(method string, err error) error {
	return fmt.Errorf("subscription usecase: %s: %w", method, err)
}
