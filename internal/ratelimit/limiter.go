package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Token Bucket をRedis Luaスクリプトでアトミックに実装。
// elapsed時間に応じてトークンを補充し、1トークン消費できれば許可する。
const luaTokenBucket = `
local key          = KEYS[1]
local rate         = tonumber(ARGV[1])
local burst        = tonumber(ARGV[2])
local now          = tonumber(ARGV[3])

local bucket       = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens       = tonumber(bucket[1])
local last_refill  = tonumber(bucket[2])

if tokens == nil then
    tokens      = burst
    last_refill = now
end

local elapsed   = math.max(0, now - last_refill) / 1000000
local new_tokens = math.min(burst, tokens + elapsed * rate)

local allowed = 0
if new_tokens >= 1 then
    new_tokens = new_tokens - 1
    allowed    = 1
end

local ttl = math.ceil(burst / rate) + 1
redis.call('HMSET', key, 'tokens', new_tokens, 'last_refill', now)
redis.call('EXPIRE', key, ttl)

return allowed
`

type Limiter struct {
	rdb    *redis.Client
	rate   float64
	burst  float64
	script *redis.Script
}

func NewLimiter(rdb *redis.Client, rps, burst int) *Limiter {
	return &Limiter{
		rdb:    rdb,
		rate:   float64(rps),
		burst:  float64(burst),
		script: redis.NewScript(luaTokenBucket),
	}
}

// Allow はIPアドレスに対してトークンを1消費し、許可するか返す。
// Redisエラー時はフェイルオープン（trueを返す）。
func (l *Limiter) Allow(ctx context.Context, ip string) (bool, error) {
	key := fmt.Sprintf("rl:%s", ip)
	now := time.Now().UnixMicro()

	result, err := l.script.Run(ctx, l.rdb, []string{key}, l.rate, l.burst, now).Int()
	if err != nil {
		return true, err // fail-open
	}
	return result == 1, nil
}
