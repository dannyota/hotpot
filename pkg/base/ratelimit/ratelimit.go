package ratelimit

import (
	"context"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

// Limiter abstracts rate limiting. Activities call Wait() before each API request.
// Implementations: local (*rate.Limiter), Redis (distributed).
type Limiter interface {
	Wait(ctx context.Context) error
}

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
	limiter Limiter
}

// NewRateLimitedTransport creates a transport that calls limiter.Wait()
// before each request. If base is nil, http.DefaultTransport is used.
func NewRateLimitedTransport(limiter Limiter, base http.RoundTripper) *RateLimitedTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &RateLimitedTransport{base: base, limiter: limiter}
}

const (
	maxRetries       = 5
	initialBackoff   = 5 * time.Second
	backoffMult      = 2.0
	maxBackoff       = 60 * time.Second
)

// RoundTrip implements http.RoundTripper with rate limiting and 429 retry.
func (t *RateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		if err := t.limiter.Wait(req.Context()); err != nil {
			return nil, err
		}

		resp, err := t.base.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests || attempt >= maxRetries {
			return resp, nil
		}

		resp.Body.Close()

		backoff := retryAfter(resp, attempt)
		slog.Warn("rate limited (429), backing off",
			"attempt", attempt+1,
			"backoff", backoff,
			"url", req.URL.Path,
		)

		select {
		case <-time.After(backoff):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}
}

// retryAfter computes the backoff duration from a 429 response.
// Uses the Retry-After header if present, otherwise exponential backoff.
func retryAfter(resp *http.Response, attempt int) time.Duration {
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if secs, err := strconv.Atoi(ra); err == nil && secs > 0 {
			return time.Duration(secs) * time.Second
		}
	}
	backoff := time.Duration(float64(initialBackoff) * math.Pow(backoffMult, float64(attempt)))
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	return backoff
}

// UnaryInterceptor returns a gRPC unary client interceptor that calls
// limiter.Wait() before each RPC. Pass as grpc.WithUnaryInterceptor().
func UnaryInterceptor(limiter Limiter) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
