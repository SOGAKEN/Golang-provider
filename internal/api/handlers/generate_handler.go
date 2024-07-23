package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"provider/internal/config"
	"provider/internal/models"
	"provider/internal/providers"
	"provider/internal/storage"

	"github.com/gin-gonic/gin"
)

func HandleGenerate(bqClient *storage.BigQueryClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 各リクエストで設定を再読み込み
		cfg, err := config.Load()
		if err != nil {
			log.Printf("Failed to load config: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		var req models.Request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		defaultProviders := cfg.GetDefaultProviders()
		responses := make(map[string]map[string][]string)

		var wg sync.WaitGroup
		var mu sync.Mutex

		jst, _ := time.LoadLocation("Asia/Tokyo")

		for _, providerName := range defaultProviders {
			wg.Add(1)
			go func(providerName string) {
				defer wg.Done()

				provider, err := providers.GetProvider(cfg, providerName)
				if err != nil {
					log.Printf("Failed to get provider %s: %v", providerName, err)
					return
				}

				prompt, err := cfg.GetProviderPrompt(providerName)
				if err != nil {
					log.Printf("Failed to get prompt for provider %s: %v", providerName, err)
					return
				}

				availableModels := provider.GetModels()
				providerResponses := make(map[string][]string)

				for _, model := range availableModels {
					if req.Parallel {
						parallelExecution(model, prompt, providerName, provider, bqClient, jst, &providerResponses)
					} else {
						sequentialExecution(model, prompt, providerName, provider, bqClient, jst, &providerResponses)
					}
				}

				mu.Lock()
				responses[providerName] = providerResponses
				mu.Unlock()
			}(providerName)
		}

		wg.Wait()

		c.JSON(http.StatusOK, responses)
	}
}

func parallelExecution(model, prompt, providerName string, provider providers.Provider, bqClient *storage.BigQueryClient, jst *time.Location, responses *map[string][]string) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	modelResponses := make([]string, 0, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(attempt int) {
			defer wg.Done()
			response := executeAndLog(model, prompt, providerName, provider, bqClient, jst, attempt)
			mu.Lock()
			modelResponses = append(modelResponses, response)
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	(*responses)[model] = modelResponses
}

func sequentialExecution(model, prompt, providerName string, provider providers.Provider, bqClient *storage.BigQueryClient, jst *time.Location, responses *map[string][]string) {
	modelResponses := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		response := executeAndLog(model, prompt, providerName, provider, bqClient, jst, i)
		modelResponses = append(modelResponses, response)
		time.Sleep(time.Duration(1+i) * time.Second) // Increasing delay between requests
	}
	(*responses)[model] = modelResponses
}

func executeAndLog(model, prompt, providerName string, provider providers.Provider, bqClient *storage.BigQueryClient, jst *time.Location, attempt int) string {
	startTime := time.Now().In(jst)
	response, err := provider.Generate(prompt, model)
	endTime := time.Now().In(jst)

	if err != nil {
		log.Printf("Generation failed for provider %s, model %s (attempt %d): %v", providerName, model, attempt+1, err)
		return fmt.Sprintf("Error: %v", err)
	}

	responseDuration := endTime.Sub(startTime)

	logEntry := &storage.GenerationLog{
		Timestamp:    startTime,
		ResponseTime: responseDuration.Seconds(),
		RequestBody:  prompt,
		ResponseBody: response,
		Model:        model,
		Provider:     providerName,
	}

	if err := bqClient.InsertGenerationLog(logEntry); err != nil {
		log.Printf("Failed to insert log to BigQuery for provider %s, model %s (attempt %d): %v", providerName, model, attempt+1, err)
	}

	log.Printf("Provider: %s, Model: %s, Attempt: %d", providerName, model, attempt+1)
	log.Printf("回答時間: %v", responseDuration)
	log.Printf("レスポンスボディ: %s", response)

	return response
}
