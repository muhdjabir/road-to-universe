package handler

import "github.com/gin-gonic/gin"

func SetupRouter(sessionHandler *SessionHandler) *gin.Engine {
	r := gin.Default()

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
			training.DELETE("/:id", sessionHandler.Delete)
		}
	}

	return r
}
