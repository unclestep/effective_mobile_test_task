package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"em/internal/postgres"
	"em/internal/subscription"

	"github.com/jackc/pgx/v5"
)

const (
	queryTimeout = 15 * time.Second
)

type pgRepo struct {
	dbtx postgres.DBTX
}

func NewPgRepo(dbtx postgres.DBTX) subscription.SubscriptionRepo {
	return &pgRepo{
		dbtx: dbtx,
	}
}

func (r *pgRepo) makeParams(f *subscription.SubscriptionFilter) (string, []any) {
	if f == nil {
		return "", nil
	}

	var sb strings.Builder
	sb.Grow(128)
	var args []any

	if f.SubID != nil {
		args = append(args, *f.SubID)
		fmt.Fprintf(&sb, " AND subscription_id = $%d", len(args))
	}
	if f.UserID != nil {
		args = append(args, *f.UserID)
		fmt.Fprintf(&sb, " AND user_id = $%d", len(args))
	}
	if f.ServiceName != nil {
		args = append(args, *f.ServiceName)
		fmt.Fprintf(&sb, " AND service_name = $%d", len(args))
	}
	// Selects subscriptions, ending after the given date (including)
	if f.PeriodFrom != nil {
		args = append(args, *f.PeriodFrom)
		fmt.Fprintf(&sb, " AND (end_date IS NULL OR end_date >= $%d)", len(args))
	}
	// Selects subscriptions, starting before the given date (including)
	if f.PeriodTo != nil {
		args = append(args, *f.PeriodTo)
		fmt.Fprintf(&sb, " AND start_date <= $%d", len(args))
	}

	return sb.String(), args
}

func (r *pgRepo) Get(parent context.Context, filter *subscription.SubscriptionFilter) ([]*subscription.Subscription, error) {
	queryParams, args := r.makeParams(filter)

	query := `
		SELECT subscription_id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE 1=1
	` + queryParams

	ctx, cancel := context.WithTimeout(parent, queryTimeout)
	defer cancel()

	rows, err := r.dbtx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("pgRepo: get: %w", err)
	}

	subscriptions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[subscription.Subscription])
	if err != nil {
		return nil, fmt.Errorf("pgRepo: get: %w", err)
	}

	return subscriptions, nil
}

func (r *pgRepo) Create(parent context.Context, sub *subscription.Subscription) error {
	query := `
		INSERT INTO subscriptions(subscription_id, service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	ctx, cancel := context.WithTimeout(parent, queryTimeout)
	defer cancel()

	_, err := r.dbtx.Exec(ctx, query, sub.SubscriptionID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate)
	if err != nil {
		return fmt.Errorf("pgRepo: create: %w", err)
	}

	return nil
}

func (r *pgRepo) Update(parent context.Context, sub *subscription.Subscription) error {
	query := `
		UPDATE subscriptions
		SET	service_name = $1,
			price = $2,
			start_date = $3,
			end_date = $4
		WHERE subscription_id = $5
	`

	ctx, cancel := context.WithTimeout(parent, queryTimeout)
	defer cancel()

	tag, err := r.dbtx.Exec(ctx, query, sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, sub.SubscriptionID)
	if err != nil {
		return fmt.Errorf("pgRepo: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("pgRepo: update: %w", subscription.ErrNotFound)
	}

	return nil
}

func (r *pgRepo) Delete(parent context.Context, subID string) error {
	query := `
		DELETE FROM subscriptions WHERE subscription_id = $1
	`

	ctx, cancel := context.WithTimeout(parent, queryTimeout)
	defer cancel()

	tag, err := r.dbtx.Exec(ctx, query, subID)
	if err != nil {
		return fmt.Errorf("pgRepo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("pgRepo: delete: %w", subscription.ErrNotFound)
	}

	return nil
}
