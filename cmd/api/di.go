package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"em/config"
	"em/internal/logger"
	"em/internal/postgres"
	"em/internal/server"
	subhttp "em/internal/subscription/delivery/http"
	"em/internal/subscription/repo"
	"em/internal/subscription/usecase"
)

func NewApp(cfg *config.Config) *fx.App {
	return fx.New(
		fx.Supply(cfg.Postgres, cfg.Server, cfg.Logger),
		fx.Provide(
			logger.New,
			postgres.NewPool,
			repo.NewPgRepo,
			usecase.NewSubscriptionUC,
			subhttp.NewHandler,
			server.New,
		),
		fx.Invoke(
			registerPoolLifecycle,
			registerRoutes,
			runHTTPServer,
		),
	)
}

func registerPoolLifecycle(lc fx.Lifecycle, pool *pgxpool.Pool, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			logger.Info("closing postgres pool")
			pool.Close()
			return nil
		},
	})
}

func registerRoutes(router *gin.Engine, handler *subhttp.Handler) {
	api := router.Group("/api/v1")
	subhttp.RegisterRoutes(api, handler)
}

func runHTTPServer(lc fx.Lifecycle, router *gin.Engine, cfg *config.ServerConfig, logger *zap.Logger) {
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting http server", zap.String("addr", srv.Addr))
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error("http server stopped unexpectedly", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("shutting down http server")
			return srv.Shutdown(ctx)
		},
	})
}
