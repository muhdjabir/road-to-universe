package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "analytics-service"})
	})

	// TODO Phase 5: consume RabbitMQ events and implement analytics
	r.GET("/api/v1/analytics/weekly", func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "not implemented — Phase 5"})
	})
	r.GET("/api/v1/analytics/stats", func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "not implemented — Phase 5"})
	})

	log.Printf("analytics service starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
