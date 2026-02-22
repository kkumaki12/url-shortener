package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/redis/go-redis/v9"

	appconfig "github.com/kumakikensuke/url-shortener/internal/config"
	"github.com/kumakikensuke/url-shortener/internal/handler"
	"github.com/kumakikensuke/url-shortener/internal/ratelimit"
	"github.com/kumakikensuke/url-shortener/internal/repository"
	"github.com/kumakikensuke/url-shortener/internal/service"
)

func main() {
	cfg := appconfig.Load()

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.AWSRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKey, cfg.AWSSecretKey, "",
		)),
	)
	if err != nil {
		log.Fatalf("load aws config: %v", err)
	}

	dynamoClient := dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
		if cfg.AWSEndpoint != "" {
			o.BaseEndpoint = aws.String(cfg.AWSEndpoint)
		}
	})

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("warning: redis ping failed: %v (rate limiter will fail-open)", err)
	}

	dynamoRepo := repository.NewDynamoRepository(dynamoClient, cfg.DynamoDBTable)
	repo := repository.NewCachedRepository(dynamoRepo, rdb, cfg.CacheTTL)
	svc := service.NewShortener(repo, cfg.BaseURL)
	limiter := ratelimit.NewLimiter(rdb, cfg.RateLimitRPS, cfg.RateLimitBurst)
	h := handler.New(svc, limiter)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := ":" + cfg.Port
	log.Printf("starting server on %s (rate limit: %d rps, burst: %d)", addr, cfg.RateLimitRPS, cfg.RateLimitBurst)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
