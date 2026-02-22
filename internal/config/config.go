package config

import "os"

type Config struct {
	Port            string
	BaseURL         string
	AWSEndpoint     string
	AWSRegion       string
	AWSAccessKey    string
	AWSSecretKey    string
	DynamoDBTable   string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		AWSEndpoint:   getEnv("AWS_ENDPOINT", "http://localhost:4566"),
		AWSRegion:     getEnv("AWS_REGION", "ap-northeast-1"),
		AWSAccessKey:  getEnv("AWS_ACCESS_KEY_ID", "dummy"),
		AWSSecretKey:  getEnv("AWS_SECRET_ACCESS_KEY", "dummy"),
		DynamoDBTable: getEnv("DYNAMODB_TABLE", "urls"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
