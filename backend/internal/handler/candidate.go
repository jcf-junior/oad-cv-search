package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"cv-search/internal/model"
	"cv-search/internal/service"
)

const (
	maxUploadSize = 10 << 20
	defaultLimit  = 5
	maxLimit      = 20
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type CandidateHandler struct {
	service *service.CandidateService
}

func NewCandidateHandler(service *service.CandidateService) *CandidateHandler {
	return &CandidateHandler{service: service}
}

func (h *CandidateHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "invalid upload")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "field 'file' is required")
		return
	}
	defer file.Close()

	if strings.ToLower(filepath.Ext(header.Filename)) != ".pdf" {
		writeError(w, http.StatusBadRequest, "only pdf files are accepted")
		return
	}

	candidate, err := h.service.Upload(r.Context(), header.Filename, file, header.Size)
	if errors.Is(err, service.ErrInvalidPDF) {
		writeError(w, http.StatusBadRequest, "invalid pdf")
		return
	}
	if errors.Is(err, service.ErrEmptyPDF) {
		writeError(w, http.StatusUnprocessableEntity, "no text found in pdf")
		return
	}
	if err != nil {
		log.Printf("upload candidate: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, candidate)
}

func (h *CandidateHandler) Count(w http.ResponseWriter, r *http.Request) {
	count, err := h.service.Count(r.Context())
	if err != nil {
		log.Printf("count candidates: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (h *CandidateHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !uuidPattern.MatchString(id) {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	candidate, err := h.service.Get(r.Context(), id)
	if errors.Is(err, model.ErrNotFound) {
		writeError(w, http.StatusNotFound, "candidate not found")
		return
	}
	if err != nil {
		log.Printf("get candidate: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, candidate)
}

type searchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func parseSearchRequest(w http.ResponseWriter, r *http.Request) (searchRequest, bool) {
	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return req, false
	}

	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return req, false
	}

	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		req.Limit = maxLimit
	}

	return req, true
}

func (h *CandidateHandler) Search(w http.ResponseWriter, r *http.Request) {
	req, ok := parseSearchRequest(w, r)
	if !ok {
		return
	}

	results, err := h.service.Search(r.Context(), req.Query, req.Limit)
	if err != nil {
		log.Printf("search candidates: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func (h *CandidateHandler) SearchKeyword(w http.ResponseWriter, r *http.Request) {
	req, ok := parseSearchRequest(w, r)
	if !ok {
		return
	}

	results, err := h.service.SearchKeyword(r.Context(), req.Query, req.Limit)
	if err != nil {
		log.Printf("keyword search candidates: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, results)
}
