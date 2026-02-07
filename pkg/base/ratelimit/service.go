package ratelimit

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"hotpot/pkg/base/config"
)

// Service manages rate limiter lifecycle.
// Created per provider in Register(). Use Limiter() to get the active limiter.
type Service struct {
	limiter     Limiter
	redisClient *redis.Client // nil if Redis not configured; held for cleanup
}

// ServiceOptions configures the rate limit Service.
type ServiceOptions struct {
	// RedisConfig enables distributed rate limiting. Nil = local only.
	RedisConfig *config.RedisConfig

	// KeyPrefix is the Redis key prefix for this provider (e.g., "ratelimit:gcp").
	KeyPrefix string

	// ReqPerMin is the rate limit in requests per minute.
	ReqPerMin int
}

// NewService creates a rate limit Service.
// If RedisConfig is provided and reachable: Redis limiter.
// If RedisConfig is provided but unreachable: local limiter (logs warning).
// If RedisConfig is nil: local limiter.
func NewService(opts ServiceOptions) *Service {
	if opts.RedisConfig == nil || opts.RedisConfig.Address == "" {
		return &Service{limiter: NewLimiter(opts.ReqPerMin)}
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     opts.RedisConfig.Address,
		Password: opts.RedisConfig.Password,
		DB:       opts.RedisConfig.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Redis unreachable (%s), falling back to local rate limiter: %v",
			opts.RedisConfig.Address, err)
		redisClient.Close()
		return &Service{limiter: NewLimiter(opts.ReqPerMin)}
	}

	log.Printf("Rate limiter using Redis at %s (key prefix: %s)", opts.RedisConfig.Address, opts.KeyPrefix)

	return &Service{
		limiter: NewRedisLimiter(RedisLimiterOptions{
			Client:    redisClient,
			KeyPrefix: opts.KeyPrefix,
			ReqPerMin: opts.ReqPerMin,
		}),
		redisClient: redisClient,
	}
}

// Limiter returns the active Limiter.
func (s *Service) Limiter() Limiter {
	return s.limiter
}

// Close releases resources (Redis connection).
func (s *Service) Close() error {
	if s.redisClient != nil {
		return s.redisClient.Close()
	}
	return nil
}
