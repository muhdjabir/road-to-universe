package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("notification worker starting...")

	// TODO Phase 6: connect to RabbitMQ, consume events, send notifications
	_ = os.Getenv("RABBITMQ_URL")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("notification worker shutting down")
}
