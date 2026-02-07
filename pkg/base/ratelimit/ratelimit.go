package ratelimit

import (
	"context"
	"net/http"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

// NewLimiter creates a rate limiter from requests-per-minute.
// Burst = max(1, reqPerMin/60) for per-second smoothing.
func NewLimiter(reqPerMin int) *rate.Limiter {
	perSec := float64(reqPerMin) / 60.0
	burst := max(1, reqPerMin/60)
	return rate.NewLimiter(rate.Limit(perSec), burst)
}

// RateLimitedTransport wraps an http.RoundTripper with rate limiting.
// Use when a SDK accepts a custom http.Client.
type RateLimitedTransport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

// NewRateLimitedTransport creates a transport that calls limiter.Wait()
// before each request. If base is nil, http.DefaultTransport is used.
func NewRateLimitedTransport(limiter *rate.Limiter, base http.RoundTripper) *RateLimitedTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &RateLimitedTransport{base: base, limiter: limiter}
}

// RoundTrip implements http.RoundTripper with rate limiting.
func (t *RateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.limiter.Wait(req.Context()); err != nil {
		return nil, err
	}
	return t.base.RoundTrip(req)
}

// UnaryInterceptor returns a gRPC unary client interceptor that calls
// limiter.Wait() before each RPC. Pass as grpc.WithUnaryInterceptor().
func UnaryInterceptor(limiter *rate.Limiter) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
