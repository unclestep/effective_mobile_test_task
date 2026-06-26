package postgres

import (
	"context"
	"fmt"

	"em/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewPool(parent context.Context, cfg *config.PostgresConfig, logger *zap.Logger) (*pgxpool.Pool, error) {
	pgCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		logger.Error("failed to parse postgres config",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("database", cfg.DBName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("new pool: %w", err)
	}

	pgCfg.MaxConns = cfg.MaxConns
	pgCfg.MinConns = cfg.MinConns
	pgCfg.MaxConnLifetime = cfg.MaxConnLifetime.Duration()
	pgCfg.MaxConnIdleTime = cfg.MaxConnIdleTime.Duration()
	pgCfg.HealthCheckPeriod = cfg.HealthCheckPeriod.Duration()
	pgCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout.Duration()

	ctx, cancel := context.WithTimeout(parent, cfg.PoolInitTimeout.Duration())
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		logger.Error("failed to create a new postgres pool",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("database", cfg.DBName),
			zap.Error(err))
		return nil, fmt.Errorf("new pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		logger.Error("failed to create a new postgres pool: didn't receive a pong",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("database", cfg.DBName),
			zap.Error(err))
		return nil, fmt.Errorf("new pool: %w", err)
	}

	return pool, nil
}
