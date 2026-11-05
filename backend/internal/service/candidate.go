package service

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"cv-search/internal/model"
	"cv-search/internal/repository"
)

type CandidateService struct {
	repo       *repository.CandidateRepository
	embeddings *EmbeddingService
}

func NewCandidateService(repo *repository.CandidateRepository, embeddings *EmbeddingService) *CandidateService {
	return &CandidateService{repo: repo, embeddings: embeddings}
}

func (s *CandidateService) Upload(ctx context.Context, fileName string, file io.ReaderAt, size int64) (model.Candidate, error) {
	content, err := ExtractText(file, size)
	if err != nil {
		return model.Candidate{}, err
	}

	embedding, err := s.embeddings.Embed(ctx, content)
	if err != nil {
		return model.Candidate{}, err
	}

	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	return s.repo.Create(ctx, name, fileName, content, embedding)
}

func (s *CandidateService) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

func (s *CandidateService) Get(ctx context.Context, id string) (model.Candidate, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CandidateService) Search(ctx context.Context, query string, limit int) ([]model.SearchResult, error) {
	embedding, err := s.embeddings.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	return s.repo.SearchByEmbedding(ctx, embedding, limit)
}

func (s *CandidateService) SearchKeyword(ctx context.Context, query string, limit int) ([]model.KeywordResult, error) {
	return s.repo.SearchByKeyword(ctx, query, limit)
}
