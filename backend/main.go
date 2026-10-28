package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"cv-search/internal/config"
	"cv-search/internal/handler"
	"cv-search/internal/ollama"
	"cv-search/internal/repository"
	"cv-search/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	ollamaClient := ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel)
	embeddingService := service.NewEmbeddingService(ollamaClient)
	candidateRepo := repository.NewCandidateRepository(pool)
	candidateService := service.NewCandidateService(candidateRepo, embeddingService)
	candidateHandler := handler.NewCandidateHandler(candidateService)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler.NewRouter(candidateHandler, cfg.FrontendDir),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
