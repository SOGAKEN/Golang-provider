package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"provider/internal/config"
	"provider/internal/models"
	"provider/internal/providers"
	"provider/internal/storage"

	"github.com/gin-gonic/gin"
)

func HandleGenerate(cfg *config.Config, bqClient *storage.BigQueryClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.Request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		prompt := `Human: 続きを教えて下さい。じゅげむじゅげむ

Assistant:`

		providerName := cfg.GetDefaultProvider()
		provider, err := providers.GetProvider(cfg, providerName)
		if err != nil {
			log.Printf("Failed to get provider: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		availableModels := provider.GetModels()
		responses := make(map[string][]string)

		jst, _ := time.LoadLocation("Asia/Tokyo")

		for _, model := range availableModels {
			responses[model] = make([]string, 0, 10)
			if req.Parallel {
				parallelExecution(model, prompt, providerName, provider, bqClient, jst, &responses)
			} else {
				sequentialExecution(model, prompt, providerName, provider, bqClient, jst, &responses)
			}
		}

		c.JSON(http.StatusOK, responses)
	}
}

func parallelExecution(model, prompt, providerName string, provider providers.Provider, bqClient *storage.BigQueryClient, jst *time.Location, responses *map[string][]string) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(attempt int) {
			defer wg.Done()

			executeAndLog(model, prompt, providerName, provider, bqClient, jst, responses, &mu, attempt)
		}(i)
	}

	wg.Wait()
}

func sequentialExecution(model, prompt, providerName string, provider providers.Provider, bqClient *storage.BigQueryClient, jst *time.Location, responses *map[string][]string) {
	var mu sync.Mutex
	for i := 0; i < 10; i++ {
		executeAndLog(model, prompt, providerName, provider, bqClient, jst, responses, &mu, i)

		if i < 9 { // 最後の実行後はスリープしない
			sleepDuration := time.Duration(rand.Intn(4000)+1000) * time.Millisecond
			time.Sleep(sleepDuration)
		}
	}
}

func executeAndLog(model, prompt, providerName string, provider providers.Provider, bqClient *storage.BigQueryClient, jst *time.Location, responses *map[string][]string, mu *sync.Mutex, attempt int) {
	modelStartTime := time.Now().In(jst)
	response, err := provider.Generate(prompt, model)
	modelEndTime := time.Now().In(jst)

	mu.Lock()
	defer mu.Unlock()

	if err != nil {
		log.Printf("Generation failed for model %s (attempt %d): %v", model, attempt+1, err)
		(*responses)[model] = append((*responses)[model], fmt.Sprintf("Error: %v", err))
	} else {
		(*responses)[model] = append((*responses)[model], response)
	}

	responseDuration := modelEndTime.Sub(modelStartTime)

	logEntry := &storage.GenerationLog{
		Timestamp:    modelStartTime,
		ResponseTime: responseDuration.Seconds(),
		RequestBody:  prompt,
		ResponseBody: response,
		Model:        model,
		Provider:     providerName,
	}

	if err := bqClient.InsertGenerationLog(logEntry); err != nil {
		log.Printf("Failed to insert log to BigQuery for model %s (attempt %d): %v", model, attempt+1, err)
	}

	log.Printf("Model: %s, Attempt: %d", model, attempt+1)
	log.Printf("回答時間: %v", responseDuration)
	log.Printf("レスポンスボディ: %s", response)
	log.Printf("プロバイダー: %s", providerName)
}

