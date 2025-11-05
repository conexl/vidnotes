// config/openrouter.go
package config

import (
	"os"
	"strconv"
)

type OpenRouterConfig struct {
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Timeout   int    `json:"timeout"`
}

func GetOpenRouterConfig() *OpenRouterConfig {
	return &OpenRouterConfig{
		APIKey:    getEnv("OPENROUTER_API_KEY", ""),
		BaseURL:   getEnv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
		Model:     getEnv("OPENROUTER_MODEL", "openai/gpt-3.5-turbo"),
		MaxTokens: getEnvInt("OPENROUTER_MAX_TOKENS", 2000),
		Timeout:   getEnvInt("OPENROUTER_TIMEOUT", 120), // Увеличили до 120 секунд
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
