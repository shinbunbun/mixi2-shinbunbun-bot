package config

import (
	"fmt"
	"os"
)

type Config struct {
	ClientID           string
	ClientSecret       string
	TokenURL           string
	APIAddress         string
	SignaturePublicKey  string
	Port               string
	HealthPort         string
	DailyPostCron      string
	GitHubToken        string
	LLMBaseURL         string
	LLMModel           string
}

func Load() (*Config, error) {
	cfg := &Config{
		ClientID:          os.Getenv("CLIENT_ID"),
		ClientSecret:      os.Getenv("CLIENT_SECRET"),
		TokenURL:          os.Getenv("TOKEN_URL"),
		APIAddress:        os.Getenv("API_ADDRESS"),
		SignaturePublicKey: os.Getenv("SIGNATURE_PUBLIC_KEY"),
		Port:              os.Getenv("PORT"),
		HealthPort:        os.Getenv("HEALTH_PORT"),
		DailyPostCron:     os.Getenv("DAILY_POST_CRON"),
		GitHubToken:       os.Getenv("GITHUB_TOKEN"),
		LLMBaseURL:        os.Getenv("LLM_BASE_URL"),
		LLMModel:          os.Getenv("LLM_MODEL"),
	}

	if cfg.ClientID == "" {
		return nil, fmt.Errorf("CLIENT_ID is required")
	}
	if cfg.ClientSecret == "" {
		return nil, fmt.Errorf("CLIENT_SECRET is required")
	}
	if cfg.TokenURL == "" {
		return nil, fmt.Errorf("TOKEN_URL is required")
	}
	if cfg.APIAddress == "" {
		return nil, fmt.Errorf("API_ADDRESS is required")
	}
	if cfg.SignaturePublicKey == "" {
		return nil, fmt.Errorf("SIGNATURE_PUBLIC_KEY is required")
	}
	if cfg.GitHubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required")
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.HealthPort == "" {
		cfg.HealthPort = "8081"
	}
	if cfg.DailyPostCron == "" {
		cfg.DailyPostCron = "0 0 * * *"
	}
	if cfg.LLMBaseURL == "" {
		cfg.LLMBaseURL = "http://192.168.1.5:8081/v1"
	}
	if cfg.LLMModel == "" {
		cfg.LLMModel = "mlx-community/Qwen3.5-4B-MLX-4bit"
	}

	return cfg, nil
}
