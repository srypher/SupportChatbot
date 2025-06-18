package openai

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
	model  string
}

func NewClient(apiKey string) *Client {
	return &Client{
		client: openai.NewClient(apiKey),
		model:  "gpt-4-turbo-preview",
	}
}

func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	resp, err := c.client.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: openai.AdaEmbeddingV2,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return resp.Data[0].Embedding, nil
}

func (c *Client) GenerateChatResponse(ctx context.Context, query string, context []string) (string, error) {
	// Prepare the system message with context
	systemMsg := "You are a helpful support assistant. Use the provided context to answer questions accurately. Always cite your sources."

	// Prepare the user message with context
	userMsg := fmt.Sprintf("Context:\n%s\n\nQuestion: %s",
		formatContext(context),
		query,
	)

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemMsg,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userMsg,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate chat response: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func formatContext(contexts []string) string {
	var formatted strings.Builder
	for i, ctx := range contexts {
		formatted.WriteString(fmt.Sprintf("[%d] %s\n", i+1, ctx))
	}
	return formatted.String()
}
