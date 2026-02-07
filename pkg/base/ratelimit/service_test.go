package ratelimit

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"

	"hotpot/pkg/base/config"
)

func TestNewService_NoRedis(t *testing.T) {
	svc := NewService(ServiceOptions{
		RedisConfig: nil,
		KeyPrefix:   "ratelimit:test",
		ReqPerMin:   600,
	})
	defer svc.Close()

	// Should return a local limiter (not nil)
	if svc.Limiter() == nil {
		t.Fatal("expected non-nil limiter")
	}

	// Should work
	if err := svc.Limiter().Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewService_EmptyRedisAddress(t *testing.T) {
	svc := NewService(ServiceOptions{
		RedisConfig: &config.RedisConfig{Address: ""},
		KeyPrefix:   "ratelimit:test",
		ReqPerMin:   600,
	})
	defer svc.Close()

	// Should fall back to local limiter
	if svc.Limiter() == nil {
		t.Fatal("expected non-nil limiter")
	}
}

func TestNewService_WithRedis(t *testing.T) {
	mr := miniredis.RunT(t)

	svc := NewService(ServiceOptions{
		RedisConfig: &config.RedisConfig{Address: mr.Addr()},
		KeyPrefix:   "ratelimit:test",
		ReqPerMin:   600,
	})
	defer svc.Close()

	if svc.Limiter() == nil {
		t.Fatal("expected non-nil limiter")
	}

	// Should use Redis limiter
	if err := svc.Limiter().Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify Redis key was created
	keys := mr.Keys()
	if len(keys) == 0 {
		t.Fatal("expected Redis key to be created")
	}
}

func TestNewService_UnreachableRedis(t *testing.T) {
	svc := NewService(ServiceOptions{
		RedisConfig: &config.RedisConfig{Address: "localhost:1"},
		KeyPrefix:   "ratelimit:test",
		ReqPerMin:   600,
	})
	defer svc.Close()

	// Should fall back to local limiter (not panic or return nil)
	if svc.Limiter() == nil {
		t.Fatal("expected non-nil limiter")
	}

	if err := svc.Limiter().Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error with fallback limiter: %v", err)
	}
}

func TestService_Close(t *testing.T) {
	// Close with no Redis should not error
	svc := NewService(ServiceOptions{
		RedisConfig: nil,
		KeyPrefix:   "ratelimit:test",
		ReqPerMin:   600,
	})
	if err := svc.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Close with Redis
	mr := miniredis.RunT(t)
	svc = NewService(ServiceOptions{
		RedisConfig: &config.RedisConfig{Address: mr.Addr()},
		KeyPrefix:   "ratelimit:test",
		ReqPerMin:   600,
	})
	if err := svc.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
