package main

import (
	"log"
	"memoriva-backend/config"
	"memoriva-backend/handlers"
	"memoriva-backend/middleware"
	"memoriva-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := services.InitDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize services
	dbService := services.NewDatabaseService(db)
	llmService := services.NewLLMService(cfg.DeepSeekAPIKey, cfg.OpenAIAPIKey)
	embeddingService := services.NewEmbeddingService(cfg.OpenAIAPIKey)
	ragService := services.NewRAGService(dbService, llmService, embeddingService)

	// Initialize queue service with 3 workers for concurrent processing
	queueService := services.NewQueueService(3, ragService, dbService)
	queueService.Start()

	// Initialize handlers with queue service and database service
	studyHandler := handlers.NewStudyHandler(queueService, dbService)

	// Setup Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint (no auth required)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "memoriva-rag-backend",
		})
	})

	// API routes with authentication
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware()) // Require authentication for all API routes
	{
		studySessions := api.Group("/study-sessions")
		{
			studySessions.POST("/process", studyHandler.ProcessStudySession)
			studySessions.GET("/:id/status", studyHandler.GetStudySessionStatus)
		}
	}

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
