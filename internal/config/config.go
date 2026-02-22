package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	BaseURL         string
	AWSEndpoint     string
	AWSRegion       string
	AWSAccessKey    string
	AWSSecretKey    string
	DynamoDBTable   string
	RedisAddr       string
	RateLimitRPS    int
	RateLimitBurst  int
	CacheTTL        time.Duration
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		BaseURL:        getEnv("BASE_URL", "http://localhost:8080"),
		AWSEndpoint:    getEnv("AWS_ENDPOINT", "http://localhost:4566"),
		AWSRegion:      getEnv("AWS_REGION", "ap-northeast-1"),
		AWSAccessKey:   getEnv("AWS_ACCESS_KEY_ID", "dummy"),
		AWSSecretKey:   getEnv("AWS_SECRET_ACCESS_KEY", "dummy"),
		DynamoDBTable:  getEnv("DYNAMODB_TABLE", "urls"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RateLimitRPS:   getEnvInt("RATE_LIMIT_RPS", 10),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 20),
		CacheTTL:       time.Duration(getEnvInt("CACHE_TTL", 86400)) * time.Second,
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}
