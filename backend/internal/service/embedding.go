package service

import (
	"context"
	"fmt"

	"cv-search/internal/ollama"
)

const EmbeddingDimensions = 1024

type EmbeddingService struct {
	client *ollama.Client
}

func NewEmbeddingService(client *ollama.Client) *EmbeddingService {
	return &EmbeddingService{client: client}
}

func (s *EmbeddingService) Embed(ctx context.Context, text string) ([]float32, error) {
	embedding, err := s.client.Embed(ctx, text)
	if err != nil {
		return nil, err
	}
	if len(embedding) != EmbeddingDimensions {
		return nil, fmt.Errorf("expected %d dimensions, got %d", EmbeddingDimensions, len(embedding))
	}
	return embedding, nil
}
