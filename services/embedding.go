package services

import (
	"context"
	"fmt"
	"memoriva-backend/models"

	"github.com/sashabaranov/go-openai"
)

type EmbeddingService struct {
	client *openai.Client
}

func NewEmbeddingService(apiKey string) *EmbeddingService {
	var client *openai.Client
	if apiKey != "" {
		client = openai.NewClient(apiKey)
	}

	return &EmbeddingService{
		client: client,
	}
}

func (s *EmbeddingService) GetCardEmbedding(card models.Flashcard) ([]float32, error) {
	if s.client == nil {
		return nil, fmt.Errorf("no embedding client available")
	}

	// Combine front and back for embedding
	text := fmt.Sprintf("%s %s", card.Front, card.Back)

	resp, err := s.client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: openai.SmallEmbedding3,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("embedding API error: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return resp.Data[0].Embedding, nil
}

func (s *EmbeddingService) GetPromptEmbedding(prompt string) ([]float32, error) {
	if s.client == nil {
		return nil, fmt.Errorf("no embedding client available")
	}

	resp, err := s.client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{prompt},
			Model: openai.SmallEmbedding3,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("embedding API error: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return resp.Data[0].Embedding, nil
}

func (s *EmbeddingService) CalculateSimilarity(embedding1, embedding2 []float32) float64 {
	if len(embedding1) != len(embedding2) {
		return 0.0
	}

	var dotProduct, norm1, norm2 float64
	for i := range embedding1 {
		dotProduct += float64(embedding1[i] * embedding2[i])
		norm1 += float64(embedding1[i] * embedding1[i])
		norm2 += float64(embedding2[i] * embedding2[i])
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	return dotProduct / (norm1 * norm2)
}
