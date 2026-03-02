package gcpauth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Scope is the OAuth2 scope for GCP API access.
const Scope = "https://www.googleapis.com/auth/cloud-platform"

// NewHTTPClient builds an *http.Client with GCP auth baked into the transport.
//
// Credential resolution:
//  1. If credJSON is non-empty, uses those credentials (production).
//  2. Otherwise falls back to ADC via google.FindDefaultCredentials, which checks
//     GOOGLE_APPLICATION_CREDENTIALS, gcloud's application_default_credentials.json,
//     and the GCE/GKE metadata server.
//
// Transport chain: http.DefaultTransport → RateLimitedTransport → oauth2.Transport → http.Client
func NewHTTPClient(ctx context.Context, credJSON []byte, limiter ratelimit.Limiter) (*http.Client, error) {
	var tokenSource oauth2.TokenSource

	if len(credJSON) > 0 {
		creds, err := google.CredentialsFromJSON(ctx, credJSON, Scope)
		if err != nil {
			return nil, fmt.Errorf("gcpauth: credentials from JSON: %w", err)
		}
		tokenSource = creds.TokenSource
	} else {
		creds, err := google.FindDefaultCredentials(ctx, Scope)
		if err != nil {
			return nil, fmt.Errorf("gcpauth: find default credentials: %w", err)
		}
		tokenSource = creds.TokenSource
	}

	transport := &oauth2.Transport{
		Source: tokenSource,
		Base:   ratelimit.NewRateLimitedTransport(limiter, nil),
	}

	return &http.Client{Transport: transport}, nil
}
