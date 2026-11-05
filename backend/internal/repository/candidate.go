package repository

import (
	"context"
	"errors"

	"cv-search/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type CandidateRepository struct {
	pool *pgxpool.Pool
}

func NewCandidateRepository(pool *pgxpool.Pool) *CandidateRepository {
	return &CandidateRepository{pool: pool}
}

func (r *CandidateRepository) Create(ctx context.Context, name, fileName, content string, embedding []float32) (model.Candidate, error) {
	candidate := model.Candidate{
		Name:     name,
		FileName: fileName,
		Content:  content,
	}

	err := r.pool.QueryRow(ctx,
		`INSERT INTO candidates (name, file_name, content, embedding)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id::text, created_at`,
		name, fileName, content, pgvector.NewVector(embedding),
	).Scan(&candidate.ID, &candidate.CreatedAt)
	if err != nil {
		return model.Candidate{}, err
	}

	return candidate, nil
}

func (r *CandidateRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM candidates`).Scan(&count)
	return count, err
}

func (r *CandidateRepository) GetByID(ctx context.Context, id string) (model.Candidate, error) {
	var candidate model.Candidate

	err := r.pool.QueryRow(ctx,
		`SELECT id::text, name, file_name, content, created_at
		 FROM candidates
		 WHERE id = $1`,
		id,
	).Scan(&candidate.ID, &candidate.Name, &candidate.FileName, &candidate.Content, &candidate.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Candidate{}, model.ErrNotFound
	}
	if err != nil {
		return model.Candidate{}, err
	}

	return candidate, nil
}

func (r *CandidateRepository) SearchByEmbedding(ctx context.Context, embedding []float32, limit int) ([]model.SearchResult, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id::text, name, file_name, content, created_at,
		        1 - (embedding <=> $1) AS similarity
		 FROM candidates
		 ORDER BY embedding <=> $1
		 LIMIT $2`,
		pgvector.NewVector(embedding), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []model.SearchResult{}
	for rows.Next() {
		var result model.SearchResult
		err := rows.Scan(
			&result.ID, &result.Name, &result.FileName, &result.Content,
			&result.CreatedAt, &result.Similarity,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

func (r *CandidateRepository) SearchByKeyword(ctx context.Context, query string, limit int) ([]model.KeywordResult, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id::text, name, file_name, content, created_at,
		        ts_rank(to_tsvector('simple', content), plainto_tsquery('simple', $1))::float8 AS rank
		 FROM candidates
		 WHERE to_tsvector('simple', content) @@ plainto_tsquery('simple', $1)
		 ORDER BY rank DESC
		 LIMIT $2`,
		query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []model.KeywordResult{}
	for rows.Next() {
		var result model.KeywordResult
		err := rows.Scan(
			&result.ID, &result.Name, &result.FileName, &result.Content,
			&result.CreatedAt, &result.Rank,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, rows.Err()
}
