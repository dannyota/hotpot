package pki

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
)

// Activities holds dependencies for PKI-level Temporal activities.
type Activities struct {
	configService *config.Service
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		limiter:       limiter,
	}
}

// DiscoverMountsParams contains parameters for mount discovery.
type DiscoverMountsParams struct {
	VaultName string
}

// DiscoverMountsResult contains the result of mount discovery.
type DiscoverMountsResult struct {
	MountPaths []string
}

// DiscoverMountsActivity is the activity function reference for workflow registration.
var DiscoverMountsActivity = (*Activities).DiscoverMounts

// mountsResponse represents the Vault /v1/sys/mounts API response.
type mountsResponse struct {
	Data map[string]mountInfo `json:"data"`
}

type mountInfo struct {
	Type string `json:"type"`
}

// DiscoverMounts calls GET /v1/sys/mounts and filters for PKI type mounts.
func (a *Activities) DiscoverMounts(ctx context.Context, params DiscoverMountsParams) (*DiscoverMountsResult, error) {
	logger := activity.GetLogger(ctx)

	inst := a.configService.VaultInstance(params.VaultName)
	if inst == nil {
		return nil, fmt.Errorf("vault instance %q not found in config", params.VaultName)
	}

	// Create HTTP client with rate limiting and TLS config
	transport := a.createTransport(inst)
	httpClient := &http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, transport),
	}

	// Call /v1/sys/mounts
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(inst.Address, "/")+"/v1/sys/mounts", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", inst.Token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discover mounts for %s: %w", params.VaultName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discover mounts for %s: status %d: %s", params.VaultName, resp.StatusCode, string(body))
	}

	var mountsResp mountsResponse
	if err := json.NewDecoder(resp.Body).Decode(&mountsResp); err != nil {
		return nil, fmt.Errorf("decode mounts response: %w", err)
	}

	// Filter for PKI mounts
	var pkiMounts []string
	for path, info := range mountsResp.Data {
		if info.Type == "pki" {
			pkiMounts = append(pkiMounts, path)
		}
	}

	logger.Info("Discovered PKI mounts",
		"vaultName", params.VaultName,
		"mountCount", len(pkiMounts),
	)

	return &DiscoverMountsResult{
		MountPaths: pkiMounts,
	}, nil
}

// createTransport creates an HTTP transport with TLS settings.
func (a *Activities) createTransport(inst *config.VaultInstance) http.RoundTripper {
	verifySSL := true
	if inst.VerifySSL != nil {
		verifySSL = *inst.VerifySSL
	}

	if verifySSL {
		return http.DefaultTransport
	}

	return &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // User explicitly set verify_ssl: false
		},
	}
}
