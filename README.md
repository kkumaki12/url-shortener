# url-shortener

A high-throughput URL shortening service written in Go, designed to handle up to 100 million records per day (365 billion over 10 years).

## Features

- **Base62 code generation** — 8-character codes using `crypto/rand` (~218 trillion combinations)
- **Collision handling** — DynamoDB conditional writes with up to 3 retries
- **Rate limiting** — Token Bucket per client IP, backed by Redis
- **Local-first development** — DynamoDB and Redis emulated via Docker Compose (LocalStack + Redis)

## Architecture

```
Client
  │
  ▼
[Go HTTP Server :8080]
  │
  ├── POST /shorten  ──► [rate limiter: Redis Token Bucket]
  │                          │
  │                          ▼ (allowed)
  │                      [service] generate Base62 code
  │                          │
  │                          ▼
  │                      [repository] DynamoDB PutItem
  │                          (attribute_not_exists — retry up to 3x on collision)
  │
  └── GET /{code}    ──► [repository] DynamoDB GetItem ──► 301 Redirect
```

## API

### `POST /shorten`

Create a shortened URL.

```
POST /shorten
Content-Type: application/json

{"url": "https://example.com/very/long/path"}
```

**Response `201 Created`**
```json
{"code": "abc12XYZ", "short_url": "http://localhost:8080/abc12XYZ"}
```

| Status | Reason |
|--------|--------|
| 400 | Missing or invalid URL |
| 429 | Rate limit exceeded |
| 500 | Internal server error |

### `GET /{code}`

Redirect to the original URL.

**Response `301 Moved Permanently`** — `Location: <original_url>`

| Status | Reason |
|--------|--------|
| 404 | Code not found |

### `GET /health`

```json
{"status": "ok"}
```

## Rate Limiting

`POST /shorten` is rate-limited per client IP using a **Token Bucket** algorithm implemented as an atomic Redis Lua script.

- Default: **10 requests/sec**, burst up to **20**
- On limit exceeded: `HTTP 429`
- On Redis failure: **fail-open** (requests pass through, error is logged)
- Client IP resolution order: `X-Forwarded-For` → `X-Real-IP` → `RemoteAddr`

## DynamoDB Table

| Attribute | Type | Role |
|-----------|------|------|
| `code` | String | Partition key (8-char Base62) |
| `original_url` | String | Original URL |
| `created_at` | String | ISO 8601 timestamp |

Billing mode: `PAY_PER_REQUEST` — auto-scales to handle write spikes.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `BASE_URL` | `http://localhost:8080` | Host used when building short URLs |
| `AWS_ENDPOINT` | `http://localhost:4566` | DynamoDB endpoint (LocalStack) |
| `AWS_REGION` | `ap-northeast-1` | AWS region |
| `AWS_ACCESS_KEY_ID` | `dummy` | AWS credentials (any value for LocalStack) |
| `AWS_SECRET_ACCESS_KEY` | `dummy` | AWS credentials (any value for LocalStack) |
| `DYNAMODB_TABLE` | `urls` | DynamoDB table name |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `RATE_LIMIT_RPS` | `10` | Token refill rate (requests/sec per IP) |
| `RATE_LIMIT_BURST` | `20` | Maximum burst size per IP |

## Getting Started

**Prerequisites:** Docker, Docker Compose, `jq` (optional, for pretty output)

```bash
# Start LocalStack (DynamoDB), Redis, and the API server
make up

# Shorten a URL
make shorten
# {"code":"abc12XYZ","short_url":"http://localhost:8080/abc12XYZ"}

# Follow the redirect
curl -v http://localhost:8080/abc12XYZ
# HTTP/1.1 301 Moved Permanently
# Location: https://example.com/very/long/path

# Check a non-existent code
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/notfound
# 404

# Health check
make health
# {"status":"ok"}

# Stop everything
make down
```

## Local Development (without Docker for the API)

```bash
# Start only LocalStack and Redis
docker compose up localstack redis -d

# Run the API with go run (requires Go 1.22+)
make dev
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make up` | Build and start all services (LocalStack + Redis + API) |
| `make down` | Stop and remove all containers |
| `make dev` | Start dependencies only, run API with `go run` |
| `make build` | Build binary to `bin/url-shortener` |
| `make test` | Run all tests |
| `make shorten` | Send a test `POST /shorten` request |
| `make health` | Send a test `GET /health` request |

## Project Structure

```
url-shortener/
├── cmd/server/main.go              # Entry point, dependency wiring
├── internal/
│   ├── config/config.go            # Environment variable loading
│   ├── handler/handler.go          # HTTP handlers + rate limit middleware
│   ├── ratelimit/limiter.go        # Token Bucket via Redis Lua script
│   ├── service/shortener.go        # Code generation and business logic
│   └── repository/dynamo.go        # DynamoDB CRUD
├── scripts/init-dynamodb.sh        # Table creation on LocalStack startup
├── docker-compose.yml              # LocalStack + Redis + API
├── Dockerfile                      # Multi-stage build (golang:1.24-alpine)
├── go.mod
└── Makefile
```

## Scale Notes

| Metric | Value |
|--------|-------|
| Code space | 62^8 ≈ 218 trillion combinations |
| Target scale | 100M writes/day × 10 years = 365B records |
| Collision probability at 365B records | ~0.084% (birthday problem) → negligible with retries |
| DynamoDB billing | On-demand (`PAY_PER_REQUEST`) — handles write spikes automatically |
