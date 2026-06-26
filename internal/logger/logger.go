package logger

import (
	"em/config"
	"go.uber.org/zap"
)

func New(cfg *config.Config) *zap.Logger {
	var logger *zap.Logger
	if cfg.Logger.Development {
		logger = zap.Must(zap.NewDevelopment())
	} else {
		logger = zap.Must(zap.NewProduction())
	}
	return logger
}
