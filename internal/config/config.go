package config

import (
	"fmt"
	"os"
	"strings"
)

// parseBoolEnv は bool を表す環境変数文字列を解釈する。
// "1", "true", "yes", "on" (大文字小文字無視) を true、それ以外/空文字は false。
func parseBoolEnv(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

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
	// LLMDisableThinking: Qwen3 系モデル等で chat template の reasoning モードを
	// 無効化する。true の場合 OpenAI 互換リクエストに
	// chat_template_kwargs={"enable_thinking":false} を付与する。
	// 既定 (false) は OpenAI / mlx-lm 等の従来挙動。
	LLMDisableThinking bool
}

func Load() (*Config, error) {
	cfg := &Config{
		ClientID:           os.Getenv("CLIENT_ID"),
		ClientSecret:       os.Getenv("CLIENT_SECRET"),
		TokenURL:           os.Getenv("TOKEN_URL"),
		APIAddress:         os.Getenv("API_ADDRESS"),
		SignaturePublicKey: os.Getenv("SIGNATURE_PUBLIC_KEY"),
		Port:               os.Getenv("PORT"),
		HealthPort:         os.Getenv("HEALTH_PORT"),
		DailyPostCron:      os.Getenv("DAILY_POST_CRON"),
		GitHubToken:        os.Getenv("GITHUB_TOKEN"),
		LLMBaseURL:         os.Getenv("LLM_BASE_URL"),
		LLMModel:           os.Getenv("LLM_MODEL"),
		LLMDisableThinking: parseBoolEnv("LLM_DISABLE_THINKING"),
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
