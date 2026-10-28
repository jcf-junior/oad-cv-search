package model

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("candidate not found")

type Candidate struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	FileName  string    `json:"file_name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type SearchResult struct {
	Candidate
	Similarity float64 `json:"similarity"`
}

type KeywordResult struct {
	Candidate
	Rank float64 `json:"rank"`
}
