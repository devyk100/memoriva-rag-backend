package config

import (
	"os"
)

type Config struct {
	DatabaseURL        string
	DeepSeekAPIKey     string
	OpenAIAPIKey       string
	Port               string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	S3BucketName       string
	CloudFrontBaseURL  string
}

func Load() *Config {
	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		DeepSeekAPIKey:     getEnv("DEEPSEEK_API_KEY", ""),
		OpenAIAPIKey:       getEnv("OPENAI_API_KEY", ""),
		Port:               getEnv("PORT", "8080"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		S3BucketName:       getEnv("S3_BUCKET_NAME", ""),
		CloudFrontBaseURL:  getEnv("CLOUDFRONT_BASE_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
