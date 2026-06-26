package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
		createContext(),
		fx.Supply(&cfg.Postgres, &cfg.Server, &cfg.Logger),
		fx.Provide(
			logger.New,
			fx.Annotate(
				postgres.NewPool,
				fx.As(new(postgres.DBTX)),
				fx.As(fx.Self()),
			),
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

func createContext() fx.Option {
	return fx.Provide(func(lc fx.Lifecycle) context.Context {
		ctx, cancel := context.WithCancel(context.Background())

		lc.Append(fx.Hook{
			OnStop: func(stopCtx context.Context) error {
				cancel()
				return nil
			},
		})

		return ctx
	})
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
	swaggerHandler := ginSwagger.WrapHandler(swaggerfiles.Handler)
	router.GET("/swagger/*any", func(c *gin.Context) {
		if any := c.Param("any"); any == "/" || any == "" {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
			return
		}
		swaggerHandler(c)
	})
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

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
