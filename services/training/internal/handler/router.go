package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/muhdjabir/road-to-universe/services/training/internal/middleware"
	"go.uber.org/zap"
)

func SetupRouter(sessionHandler *SessionHandler, log *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestLogger(log))
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "training-service"})
	})

	v1 := r.Group("/api/v1")
	{
		training := v1.Group("/training")
		{
			training.POST("", sessionHandler.Create)
			training.GET("", sessionHandler.List)
			training.GET("/:id", sessionHandler.Get)
			training.PUT("/:id", sessionHandler.Update)
			training.DELETE("/:id", sessionHandler.Delete)
		}
	}

	return r
}
