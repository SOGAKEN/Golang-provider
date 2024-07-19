package api

import (
	"github.com/gin-gonic/gin"
	"provider/internal/api/handlers"
	"provider/internal/config"
	"provider/internal/storage"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config, bqClient *storage.BigQueryClient) {
	r.POST("/generate", handlers.HandleGenerate(cfg, bqClient))
}
