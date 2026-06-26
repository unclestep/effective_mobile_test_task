package repo

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	em "em"
	"em/config"
	pg "em/internal/postgres"
	"em/internal/subscription"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
)

const (
	testDBName = "testdb"
	testDBUser = "dbuser"
	testDBPass = "dbpassword"
)

type PgRepoSuite struct {
	suite.Suite
	container *postgres.PostgresContainer
	repo      subscription.SubscriptionRepo
	pool      *pgxpool.Pool
	tx        pgx.Tx
}

func (s *PgRepoSuite) SetupSuite() {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPass),
		postgres.BasicWaitStrategies(),
	)

	s.Require().NoError(err)
	s.container = container

	databaseURL, err := container.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	err = runMigrations(databaseURL)
	s.Require().NoError(err)

	cfg, err := newTestPostgresConfig(ctx, container)
	s.Require().NoError(err)

	pool, err := pg.NewPool(ctx, cfg, zap.NewNop())
	s.Require().NoError(err)
	s.pool = pool
}

func (s *PgRepoSuite) TearDownSuite() {
	s.pool.Close()
	err := testcontainers.TerminateContainer(s.container)
	s.Require().NoError(err)
}

func (s *PgRepoSuite) SetupTest() {
	tx, err := s.pool.Begin(context.Background())
	s.Require().NoError(err)
	s.tx = tx
	s.repo = NewPgRepo(tx)
}

func (s *PgRepoSuite) TearDownTest() {
	err := s.tx.Rollback(context.Background())
	s.Require().NoError(err)
}

func newTestPostgresConfig(ctx context.Context, container *postgres.PostgresContainer) (*config.PostgresConfig, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, fmt.Errorf("get container port: %w", err)
	}

	return &config.PostgresConfig{
		Host:              host,
		Port:              int(mappedPort.Num()),
		User:              testDBUser,
		Password:          testDBPass,
		DBName:            testDBName,
		SSLMode:           "disable",
		MaxConns:          5,
		MinConns:          1,
		MaxConnLifetime:   config.Duration(time.Hour),
		MaxConnIdleTime:   config.Duration(30 * time.Minute),
		HealthCheckPeriod: config.Duration(time.Minute),
		ConnectTimeout:    config.Duration(5 * time.Second),
		PoolInitTimeout:   config.Duration(15 * time.Second),
	}, nil
}

func runMigrations(dsn string) error {
	src, err := iofs.New(em.MigrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return fmt.Errorf("create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func newTestSubscription(opts ...func(*subscription.Subscription)) *subscription.Subscription {
	sub := &subscription.Subscription{
		SubscriptionID: uuid.NewString(),
		ServiceName:    "Yandex Plus",
		Price:          400,
		UserID:         uuid.NewString(),
		StartDate:      time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	for _, opt := range opts {
		opt(sub)
	}
	return sub
}

func TestPgRepoSuite(t *testing.T) {
	suite.Run(t, new(PgRepoSuite))
}

func (s *PgRepoSuite) TestCreateInsertsRecord() {
	ctx := context.Background()
	sub := newTestSubscription()

	s.Require().NoError(s.repo.Create(ctx, sub))

	got, err := s.repo.Get(ctx, &subscription.SubscriptionFilter{SubID: &sub.SubscriptionID})
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(sub.ServiceName, got[0].ServiceName)
	s.Equal(sub.Price, got[0].Price)
	s.Equal(sub.UserID, got[0].UserID)
	s.True(sub.StartDate.Equal(got[0].StartDate))
	s.Nil(got[0].EndDate)
}

func (s *PgRepoSuite) TestCreateDuplicateIDReturnsError() {
	ctx := context.Background()
	sub := newTestSubscription()
	s.Require().NoError(s.repo.Create(ctx, sub))

	duplicate := newTestSubscription()
	duplicate.SubscriptionID = sub.SubscriptionID

	err := s.repo.Create(ctx, duplicate)
	s.Require().Error(err)
}

func (s *PgRepoSuite) TestGetWithoutFilterReturnsAllRecords() {
	ctx := context.Background()
	first := newTestSubscription()
	second := newTestSubscription()

	s.Require().NoError(s.repo.Create(ctx, first))
	s.Require().NoError(s.repo.Create(ctx, second))

	got, err := s.repo.Get(ctx, nil)
	s.Require().NoError(err)
	s.Require().Len(got, 2)

	ids := []string{got[0].SubscriptionID, got[1].SubscriptionID}
	s.Contains(ids, first.SubscriptionID)
	s.Contains(ids, second.SubscriptionID)
}

func (s *PgRepoSuite) TestGetFiltersByUserID() {
	ctx := context.Background()
	userA := uuid.NewString()
	userB := uuid.NewString()

	subA := newTestSubscription(func(sub *subscription.Subscription) { sub.UserID = userA })
	subB := newTestSubscription(func(sub *subscription.Subscription) { sub.UserID = userB })

	s.Require().NoError(s.repo.Create(ctx, subA))
	s.Require().NoError(s.repo.Create(ctx, subB))

	got, err := s.repo.Get(ctx, &subscription.SubscriptionFilter{UserID: &userA})
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(subA.SubscriptionID, got[0].SubscriptionID)
}

func (s *PgRepoSuite) TestGetFiltersByServiceName() {
	ctx := context.Background()
	serviceName := "Spotify"

	matching := newTestSubscription(func(sub *subscription.Subscription) { sub.ServiceName = serviceName })
	other := newTestSubscription(func(sub *subscription.Subscription) { sub.ServiceName = "Yandex Plus" })

	s.Require().NoError(s.repo.Create(ctx, matching))
	s.Require().NoError(s.repo.Create(ctx, other))

	got, err := s.repo.Get(ctx, &subscription.SubscriptionFilter{ServiceName: &serviceName})
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(matching.SubscriptionID, got[0].SubscriptionID)
}

func (s *PgRepoSuite) TestGetFiltersByPeriod() {
	ctx := context.Background()

	active := newTestSubscription(func(sub *subscription.Subscription) {
		sub.StartDate = time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	})

	ended := newTestSubscription(func(sub *subscription.Subscription) {
		sub.StartDate = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
		sub.EndDate = &end
	})

	s.Require().NoError(s.repo.Create(ctx, active))
	s.Require().NoError(s.repo.Create(ctx, ended))

	periodFrom := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	periodTo := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)

	got, err := s.repo.Get(ctx, &subscription.SubscriptionFilter{
		PeriodFrom: &periodFrom,
		PeriodTo:   &periodTo,
	})

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(active.SubscriptionID, got[0].SubscriptionID)
}

func (s *PgRepoSuite) TestGetWithNoMatchesReturnsEmptySlice() {
	ctx := context.Background()
	missingUserID := uuid.NewString()

	got, err := s.repo.Get(ctx, &subscription.SubscriptionFilter{UserID: &missingUserID})

	s.Require().NoError(err)
	s.Empty(got)
}

func (s *PgRepoSuite) TestUpdateModifiesFieldsAndGetConfirms() {
	ctx := context.Background()
	sub := newTestSubscription()
	s.Require().NoError(s.repo.Create(ctx, sub))

	sub.ServiceName = "Spotify"
	sub.Price = 599
	newEnd := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	sub.EndDate = &newEnd

	s.Require().NoError(s.repo.Update(ctx, sub))

	got, err := s.repo.Get(ctx, &subscription.SubscriptionFilter{SubID: &sub.SubscriptionID})
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("Spotify", got[0].ServiceName)
	s.Equal(599, got[0].Price)
	s.Require().NotNil(got[0].EndDate)
	s.True(newEnd.Equal(*got[0].EndDate))
}

func (s *PgRepoSuite) TestUpdateNonexistentIDReturnsErrNotFound() {
	ctx := context.Background()
	sub := newTestSubscription()

	err := s.repo.Update(ctx, sub)

	s.Require().ErrorIs(err, subscription.ErrNotFound)
}

func (s *PgRepoSuite) TestDeleteRemovesRecordAndGetConfirms() {
	ctx := context.Background()
	sub := newTestSubscription()
	s.Require().NoError(s.repo.Create(ctx, sub))

	s.Require().NoError(s.repo.Delete(ctx, sub.SubscriptionID))

	got, err := s.repo.Get(ctx, &subscription.SubscriptionFilter{SubID: &sub.SubscriptionID})
	s.Require().NoError(err)
	s.Empty(got)
}

func (s *PgRepoSuite) TestDeleteNonexistentIDReturnsErrNotFound() {
	ctx := context.Background()

	err := s.repo.Delete(ctx, uuid.NewString())

	s.Require().ErrorIs(err, subscription.ErrNotFound)
}

func (s *PgRepoSuite) TestGetTotalCostSumsMatchingRecords() {
	ctx := context.Background()
	userID := uuid.NewString()

	first := newTestSubscription(func(sub *subscription.Subscription) {
		sub.UserID = userID
		sub.Price = 400
	})
	second := newTestSubscription(func(sub *subscription.Subscription) {
		sub.UserID = userID
		sub.Price = 199
	})
	other := newTestSubscription(func(sub *subscription.Subscription) {
		sub.Price = 999
	})

	s.Require().NoError(s.repo.Create(ctx, first))
	s.Require().NoError(s.repo.Create(ctx, second))
	s.Require().NoError(s.repo.Create(ctx, other))

	total, err := s.repo.GetTotalCost(ctx, &subscription.SubscriptionFilter{UserID: &userID})

	s.Require().NoError(err)
	s.Equal(599, total)
}

func (s *PgRepoSuite) TestGetTotalCostWithNoMatchesReturnsZero() {
	ctx := context.Background()
	missingUserID := uuid.NewString()

	total, err := s.repo.GetTotalCost(ctx, &subscription.SubscriptionFilter{UserID: &missingUserID})

	s.Require().NoError(err)
	s.Equal(0, total)
}
