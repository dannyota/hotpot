package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedisLimiter(t *testing.T, reqPerMin int) (*RedisLimiter, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	limiter := NewRedisLimiter(RedisLimiterOptions{
		Client:    client,
		KeyPrefix: "ratelimit:test",
		ReqPerMin: reqPerMin,
	})
	return limiter, mr
}

func TestRedisLimiter_AllowsUpToLimit(t *testing.T) {
	// 600 rpm = 10 per second
	limiter, _ := newTestRedisLimiter(t, 600)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
	}
}

func TestRedisLimiter_BlocksOverLimit(t *testing.T) {
	// 60 rpm = 1 per second
	limiter, _ := newTestRedisLimiter(t, 60)
	ctx := context.Background()

	// First request should succeed
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("first request: unexpected error: %v", err)
	}

	// Second request should block — use a short timeout to verify
	ctx2, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx2)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !isContextErr(err) {
		t.Fatalf("expected context error, got: %v", err)
	}
}

func TestRedisLimiter_ConcurrentAccess(t *testing.T) {
	// 600 rpm = 10 per second
	limiter, _ := newTestRedisLimiter(t, 600)
	ctx := context.Background()

	var granted atomic.Int64
	var wg sync.WaitGroup

	// Launch 20 goroutines, each trying one request
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx2, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
			defer cancel()
			if err := limiter.Wait(ctx2); err == nil {
				granted.Add(1)
			}
		}()
	}
	wg.Wait()

	// Should grant exactly 10 (the per-second limit)
	got := granted.Load()
	if got != 10 {
		t.Errorf("expected 10 granted, got %d", got)
	}
}

func TestRedisLimiter_PerSecondMinimum(t *testing.T) {
	// 30 rpm → 30/60 = 0, but minimum should be 1 per second
	limiter, _ := newTestRedisLimiter(t, 30)
	ctx := context.Background()

	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("first request: unexpected error: %v", err)
	}

	// Second should block
	ctx2, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	if err := limiter.Wait(ctx2); err == nil {
		t.Fatal("expected timeout for second request")
	}
}

func isContextErr(err error) bool {
	return err == context.DeadlineExceeded || err == context.Canceled
}
