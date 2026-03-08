package main

import (
	"log"

	"github.com/muhdjabir/road-to-universe/services/training/internal/config"
	"github.com/muhdjabir/road-to-universe/services/training/internal/db"
	"github.com/muhdjabir/road-to-universe/services/training/internal/handler"
	"github.com/muhdjabir/road-to-universe/services/training/internal/repository"
	"github.com/muhdjabir/road-to-universe/services/training/internal/service"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("connected to database")

	sessionRepo := repository.NewSessionRepository(database)
	sessionSvc := service.NewSessionService(sessionRepo)
	sessionHandler := handler.NewSessionHandler(sessionSvc)

	router := handler.SetupRouter(sessionHandler)

	log.Printf("training service starting on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
