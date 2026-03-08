package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "auth-service"})
	})

	// TODO Phase 2: implement auth endpoints
	r.POST("/api/v1/auth/register", func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "not implemented — Phase 2"})
	})
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "not implemented — Phase 2"})
	})

	log.Printf("auth service starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
