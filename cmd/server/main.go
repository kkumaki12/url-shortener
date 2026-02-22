package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	appconfig "github.com/kumakikensuke/url-shortener/internal/config"
	"github.com/kumakikensuke/url-shortener/internal/handler"
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

	repo := repository.NewDynamoRepository(dynamoClient, cfg.DynamoDBTable)
	svc := service.NewShortener(repo, cfg.BaseURL)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := ":" + cfg.Port
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
