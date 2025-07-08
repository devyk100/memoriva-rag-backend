package config

import (
	"os"
)

type Config struct {
	DatabaseURL    string
	DeepSeekAPIKey string
	OpenAIAPIKey   string
	Port           string
}

func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		DeepSeekAPIKey: getEnv("DEEPSEEK_API_KEY", ""),
		OpenAIAPIKey:   getEnv("OPENAI_API_KEY", ""),
		Port:           getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
