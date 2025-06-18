package vectorstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type VectorStore struct {
	client         *http.Client
	baseURL        string
	collectionName string
}

func NewVectorStore(host string, port int, collectionName string) *VectorStore {
	return &VectorStore{
		client:         &http.Client{},
		baseURL:        fmt.Sprintf("http://%s:%d", host, port),
		collectionName: collectionName,
	}
}

func (vs *VectorStore) StoreEmbedding(ctx context.Context, id string, embedding []float32, metadata map[string]interface{}) error {
	point := map[string]interface{}{
		"id":      id,
		"vector":  embedding,
		"payload": metadata,
	}

	body, err := json.Marshal(map[string]interface{}{
		"points": []interface{}{point},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal point: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/points", vs.baseURL, vs.collectionName)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := vs.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float32                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

func (vs *VectorStore) SearchSimilar(ctx context.Context, embedding []float32, limit int) ([]SearchResult, error) {
	body, err := json.Marshal(map[string]interface{}{
		"vector":       embedding,
		"limit":        limit,
		"with_payload": true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/points/search", vs.baseURL, vs.collectionName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := vs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Result []SearchResult `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Result, nil
}
