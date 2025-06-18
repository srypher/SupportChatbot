package config

import (
	"fmt"
	"os"
)

type Config struct {
	OpenAIAPIKey string
	QdrantHost   string
	QdrantPort   int
	UploadDir    string
	ChunkSize    int
	ChunkOverlap int
}

func LoadConfig() (*Config, error) {
	config := &Config{
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		QdrantHost:   getEnvOrDefault("QDRANT_HOST", "localhost"),
		QdrantPort:   6333, // Default Qdrant port
		UploadDir:    getEnvOrDefault("UPLOAD_DIR", "uploads"),
		ChunkSize:    1000, // Default chunk size in words
		ChunkOverlap: 200,  // Default overlap in words
	}

	if config.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
