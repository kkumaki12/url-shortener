FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /url-shortener ./cmd/server

FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /url-shortener /url-shortener

USER appuser

EXPOSE 8080

ENTRYPOINT ["/url-shortener"]
