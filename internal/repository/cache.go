package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// inner はキャッシュが委譲する下位リポジトリのインターフェース。
// service.Repo と同一シグネチャなので DynamoRepository をそのまま渡せる。
type inner interface {
	Put(ctx context.Context, code, originalURL string) error
	Get(ctx context.Context, code string) (*URLItem, error)
}

// CachedRepository は inner をラップし、Get に Redis read-through キャッシュを追加するデコレーター。
type CachedRepository struct {
	repo inner
	rdb  *redis.Client
	ttl  time.Duration
}

func NewCachedRepository(repo inner, rdb *redis.Client, ttl time.Duration) *CachedRepository {
	return &CachedRepository{repo: repo, rdb: rdb, ttl: ttl}
}

// Put はキャッシュを経由せず下位リポジトリに委譲する。
func (c *CachedRepository) Put(ctx context.Context, code, originalURL string) error {
	return c.repo.Put(ctx, code, originalURL)
}

// Get はまず Redis を参照し、キャッシュミス時のみ DynamoDB を叩いてキャッシュに書く。
// Redis 障害時はフェイルオープンで DynamoDB に直接アクセスする。
func (c *CachedRepository) Get(ctx context.Context, code string) (*URLItem, error) {
	key := fmt.Sprintf("url:%s", code)

	val, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		return &URLItem{Code: code, OriginalURL: val}, nil
	}
	if !errors.Is(err, redis.Nil) {
		log.Printf("cache get error (fail-open): %v", err)
	}

	// キャッシュミス — DynamoDB から取得
	item, err := c.repo.Get(ctx, code)
	if err != nil {
		return nil, err
	}

	// ベストエフォートでキャッシュに書く（エラーは無視）
	if setErr := c.rdb.Set(ctx, key, item.OriginalURL, c.ttl).Err(); setErr != nil {
		log.Printf("cache set error: %v", setErr)
	}

	return item, nil
}
