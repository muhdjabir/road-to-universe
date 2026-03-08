package main

import (
	"github.com/muhdjabir/road-to-universe/services/training/internal/config"
	"github.com/muhdjabir/road-to-universe/services/training/internal/db"
	"github.com/muhdjabir/road-to-universe/services/training/internal/handler"
	"github.com/muhdjabir/road-to-universe/services/training/internal/logger"
	"github.com/muhdjabir/road-to-universe/services/training/internal/repository"
	"github.com/muhdjabir/road-to-universe/services/training/internal/service"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	log, err := logger.New(cfg.Env)
	if err != nil {
		panic("failed to initialise logger: " + err.Error())
	}
	defer log.Sync() //nolint:errcheck

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer database.Close()
	log.Info("connected to database")

	sessionRepo := repository.NewSessionRepository(database)
	sessionSvc := service.NewSessionService(sessionRepo, log)
	sessionHandler := handler.NewSessionHandler(sessionSvc, log)

	router := handler.SetupRouter(sessionHandler, log)

	log.Info("training service starting", zap.String("port", cfg.Port), zap.String("env", cfg.Env))
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("server failed", zap.Error(err))
	}
}
