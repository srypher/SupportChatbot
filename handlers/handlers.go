package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/bduff/SupportChatbot/config"
	"github.com/bduff/SupportChatbot/openai"
	"github.com/bduff/SupportChatbot/processor"
	"github.com/bduff/SupportChatbot/vectorstore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handlers struct {
	openaiClient *openai.Client
	vectorStore  *vectorstore.VectorStore
	docProcessor *processor.DocumentProcessor
	config       *config.Config
}

func NewHandlers(
	openaiClient *openai.Client,
	vectorStore *vectorstore.VectorStore,
	docProcessor *processor.DocumentProcessor,
	config *config.Config,
) *Handlers {
	return &Handlers{
		openaiClient: openaiClient,
		vectorStore:  vectorStore,
		docProcessor: docProcessor,
		config:       config,
	}
}

func (h *Handlers) HandleFileUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filepath := filepath.Join(h.config.UploadDir, filename)

	// Save file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Process file and store in vector DB
	chunks, err := h.docProcessor.ProcessFile(c.Request.Context(), filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
		return
	}

	// Store chunks in vector DB
	for i, chunk := range chunks {
		embedding, err := h.openaiClient.GenerateEmbedding(c.Request.Context(), chunk)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate embedding: %v", err)})
			return
		}

		metadata := map[string]interface{}{
			"filename":    filename,
			"chunk_index": i,
			"text":        chunk,
		}

		if err := h.vectorStore.StoreEmbedding(c.Request.Context(), fmt.Sprintf("%s_%d", filename, i), embedding, metadata); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to store embedding: %v", err)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File processed and stored successfully",
		"filename": filename,
		"chunks":   len(chunks),
	})
}

func (h *Handlers) HandleChat(c *gin.Context) {
	var request struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Generate embedding for the query
	queryEmbedding, err := h.openaiClient.GenerateEmbedding(c.Request.Context(), request.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate query embedding: %v", err)})
		return
	}

	// Search for relevant context
	results, err := h.vectorStore.SearchSimilar(c.Request.Context(), queryEmbedding, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to search vector store: %v", err)})
		return
	}

	// Extract context from results
	var context []string
	for _, result := range results {
		if text, ok := result.Payload["text"].(string); ok {
			context = append(context, text)
		}
	}

	// Generate response using OpenAI
	response, err := h.openaiClient.GenerateChatResponse(c.Request.Context(), request.Message, context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate response: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}
