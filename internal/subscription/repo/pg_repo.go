package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"em/internal/subscription"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type dbtx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

const (
	queryTimeout = 15 * time.Second
)

type pgRepo struct {
	dbtx dbtx
}

func NewPgRepo(dbtx dbtx) subscription.SubscriptionRepo {
	return &pgRepo{
		dbtx: dbtx,
	}
}

func (r *pgRepo) makeParams(opts ...subscription.SubscriptionGetOpt) (string, []any) {
	var cfg subscription.SubscriptionGetConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	var sb strings.Builder
	sb.Grow(128)
	var args []any

	if cfg.SubID != nil {
		args = append(args, *cfg.SubID)
		fmt.Fprintf(&sb, " AND subscription_id = $%d", len(args))
	}

	if cfg.UserID != nil {
		args = append(args, *cfg.UserID)
		fmt.Fprintf(&sb, " AND user_id = $%d", len(args))
	}

	if cfg.ServiceName != nil {
		args = append(args, *cfg.ServiceName)
		fmt.Fprintf(&sb, " AND service_name = $%d", len(args))
	}

	// Selects subscriptions, ending after the given date (including)
	if cfg.StartPeriod != nil {
		args = append(args, *cfg.StartPeriod)
		fmt.Fprintf(&sb, " AND (end_date IS NULL OR end_date >= $%d)", len(args))
	}

	// Selects subscriptions, starting before the given date (including)
	if cfg.EndPeriod != nil {
		args = append(args, *cfg.EndPeriod)
		fmt.Fprintf(&sb, " AND start_date <= $%d", len(args))
	}

	return sb.String(), args
}

func (r *pgRepo) Get(parent context.Context, opts ...subscription.SubscriptionGetOpt) ([]*subscription.Subscription, error) {
	queryParams, args := r.makeParams(opts...)

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
