package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bduff/SupportChatbot/config"
	"github.com/bduff/SupportChatbot/handlers"
	"github.com/bduff/SupportChatbot/openai"
	"github.com/bduff/SupportChatbot/processor"
	"github.com/bduff/SupportChatbot/vectorstore"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize OpenAI client
	openaiClient := openai.NewClient(cfg.OpenAIAPIKey)

	// Initialize Qdrant client
	vectorStore := vectorstore.NewVectorStore(cfg.QdrantHost, cfg.QdrantPort, "support_docs")

	// Initialize document processor
	docProcessor := processor.NewDocumentProcessor(cfg.ChunkSize, cfg.ChunkOverlap)

	// Initialize handlers with dependencies
	h := handlers.NewHandlers(openaiClient, vectorStore, docProcessor, cfg)

	// Set up Gin router
	r := gin.Default()

	// Enable CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes
	r.POST("/upload", h.HandleFileUpload)
	r.POST("/chat", h.HandleChat)

	// Start server in a goroutine
	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
