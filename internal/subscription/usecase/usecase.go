package usecase

import (
	"context"
	"fmt"

	"em/internal/subscription"

	"github.com/google/uuid"
)

const (
	getMethod          = "get"
	getTotalCostMethod = "get total cost"
	createMethod       = "create"
	updateMethod       = "update"
	deleteMethod       = "delete"
)

type SubscriptionUC struct {
	repo subscription.SubscriptionRepo
}

func NewSubscriptionUC(repo subscription.SubscriptionRepo) subscription.UseCase {
	return &SubscriptionUC{
		repo: repo,
	}
}

func (u *SubscriptionUC) Get(ctx context.Context, opts ...subscription.SubscriptionGetOpt) ([]*subscription.Subscription, error) {
	subs, err := u.repo.Get(ctx, opts...)
	if err != nil {
		return nil, wrap(getMethod, err)
	}
	return subs, nil
}

func (u *SubscriptionUC) GetTotalCost(ctx context.Context, opts ...subscription.SubscriptionGetOpt) (int, error) {
	subs, err := u.repo.Get(ctx, opts...)
	if err != nil {
		return 0, wrap(getTotalCostMethod, err)
	}

	var total int
	for _, s := range subs {
		total += s.Price
	}
	return total, nil
}

func (u *SubscriptionUC) Create(ctx context.Context, subInp *subscription.CreateSubInput) error {
	sub := subscription.NewSubscription(
		uuid.NewString(), subInp.ServiceName, subInp.Price, subInp.UserID, subInp.StartDate, subInp.EndDate,
	)
	if err := u.repo.Create(ctx, sub); err != nil {
		return wrap(createMethod, err)
	}
	return nil
}

const updateUC = "update"

func (u *SubscriptionUC) Update(ctx context.Context, subID string, subInp *subscription.UpdateSubInput) error {
	if subInp == nil || (subInp.ServiceName == nil && subInp.Price == nil && subInp.StartDate == nil && subInp.EndDate == nil) {
		return wrap(updateUC, subscription.ErrNothingToUpdate)
	}

	repoSubs, err := u.repo.Get(ctx, subscription.WithSubscriptionID(subID))
	if err != nil {
		return wrap(updateUC, err)
	}
	if len(repoSubs) == 0 {
		return wrap(updateUC, subscription.ErrNotFound)
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
		return wrap(updateUC, err)
	}

	return nil
}

func (u *SubscriptionUC) Delete(ctx context.Context, subID string) error {
	if err := u.repo.Delete(ctx, subID); err != nil {
		return wrap(deleteMethod, err)
	}
	return nil
}

func wrap(method string, err error) error {
	return fmt.Errorf("subscription usecase: %s: %w", method, err)
}
