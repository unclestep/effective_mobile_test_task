package logger

import (
	"em/config"
	"go.uber.org/zap"
)

func New(cfg *config.LoggerConfig) *zap.Logger {
	var logger *zap.Logger
	if cfg.Development {
		logger = zap.Must(zap.NewDevelopment())
	} else {
		logger = zap.Must(zap.NewProduction())
	}
	return logger
}
