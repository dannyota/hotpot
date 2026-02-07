package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter implements Limiter using a per-second INCR counter in Redis.
// All workers share the same Redis keys, enforcing a global rate limit.
type RedisLimiter struct {
	client    *redis.Client
	keyPrefix string  // e.g., "ratelimit:gcp"
	perSecond int     // max requests per second (rate_limit_per_minute / 60)
}

// RedisLimiterOptions configures the Redis rate limiter.
type RedisLimiterOptions struct {
	Client    *redis.Client
	KeyPrefix string
	ReqPerMin int
}

// NewRedisLimiter creates a Redis-backed rate limiter.
func NewRedisLimiter(opts RedisLimiterOptions) *RedisLimiter {
	perSecond := opts.ReqPerMin / 60
	if perSecond < 1 {
		perSecond = 1
	}
	return &RedisLimiter{
		client:    opts.Client,
		keyPrefix: opts.KeyPrefix,
		perSecond: perSecond,
	}
}

// Wait blocks until the rate limit allows one request or ctx is cancelled.
// Uses atomic INCR on per-second keys (e.g., "ratelimit:gcp:1707300000").
func (r *RedisLimiter) Wait(ctx context.Context) error {
	for {
		sec := time.Now().Unix()
		key := fmt.Sprintf("%s:%d", r.keyPrefix, sec)

		count, err := r.client.Incr(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("redis rate limit: %w", err)
		}
		if count == 1 {
			r.client.Expire(ctx, key, 2*time.Second)
		}
		if count <= int64(r.perSecond) {
			return nil
		}

		// Over budget this second — wait for next
		next := time.Unix(sec+1, 0)
		select {
		case <-time.After(time.Until(next)):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
