package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	OllamaURL   string
	OllamaModel string
	FrontendDir string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://cvsearch:cvsearch@localhost:5432/cvsearch?sslmode=disable"),
		OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel: getEnv("OLLAMA_MODEL", "bge-m3"),
		FrontendDir: getEnv("FRONTEND_DIR", "../frontend"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
