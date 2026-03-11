package ratelimit

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
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
// Network errors automatically trigger an IPv4 fallback retry.
// Use when a SDK accepts a custom http.Client.
type RateLimitedTransport struct {
	base     http.RoundTripper
	limiter  Limiter
	ipv4Once sync.Once
	ipv4     *http.Transport
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

// RoundTrip implements http.RoundTripper with rate limiting, 429 retry,
// and IPv4 fallback on network errors.
func (t *RateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		if err := t.limiter.Wait(req.Context()); err != nil {
			return nil, err
		}

		resp, err := t.roundTripWithIPv4Fallback(req)
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

// roundTripWithIPv4Fallback tries the base transport first. On network errors
// (e.g. IPv6 connection reset by peer), it retries with an IPv4-only transport.
func (t *RateLimitedTransport) roundTripWithIPv4Fallback(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err == nil {
		return resp, nil
	}

	var netErr *net.OpError
	if !errors.As(err, &netErr) {
		return nil, err
	}

	slog.Warn("request failed, retrying with IPv4",
		"error", err,
		"url", req.URL.String(),
	)

	t.ipv4Once.Do(func() {
		t.ipv4 = &http.Transport{
			DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
				return (&net.Dialer{Timeout: 30 * time.Second}).DialContext(ctx, "tcp4", addr)
			},
			TLSHandshakeTimeout: 10 * time.Second,
		}
	})

	return t.ipv4.RoundTrip(req)
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
