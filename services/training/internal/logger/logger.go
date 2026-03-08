package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a zap logger appropriate for the environment.
// Development: coloured, human-readable, DEBUG level.
// Production:  JSON, INFO level.
func New(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return cfg.Build()
}
