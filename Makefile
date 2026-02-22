.PHONY: up down dev build test shorten health

up:
	docker compose up --build -d

down:
	docker compose down

dev:
	docker compose up localstack -d
	AWS_ENDPOINT=http://localhost:4566 \
	AWS_REGION=ap-northeast-1 \
	AWS_ACCESS_KEY_ID=dummy \
	AWS_SECRET_ACCESS_KEY=dummy \
	DYNAMODB_TABLE=urls \
	BASE_URL=http://localhost:8080 \
	go run ./cmd/server

build:
	go build -o bin/url-shortener ./cmd/server

test:
	go test ./...

shorten:
	curl -s -X POST http://localhost:8080/shorten \
		-H "Content-Type: application/json" \
		-d '{"url":"https://example.com/very/long/path"}' | jq .

health:
	curl -s http://localhost:8080/health | jq .
