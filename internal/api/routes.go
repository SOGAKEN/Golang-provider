package api

import (
	"github.com/gin-gonic/gin"
	"provider/internal/api/handlers"
	"provider/internal/storage"
)

func SetupRoutes(r *gin.Engine, bqClient *storage.BigQueryClient) {
	r.POST("/generate", handlers.HandleGenerate(bqClient))
}
