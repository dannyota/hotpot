package certificate

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Client is an HTTP client for Vault PKI certificate operations.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Vault PKI client.
func NewClient(address, token string, verifySSL bool, limiter ratelimit.Limiter) *Client {
	var transport http.RoundTripper
	if verifySSL {
		transport = http.DefaultTransport
	} else {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // User explicitly set verify_ssl: false
			},
		}
	}

	return &Client{
		baseURL: strings.TrimRight(address, "/"),
		token:   token,
		httpClient: &http.Client{
			Transport: ratelimit.NewRateLimitedTransport(limiter, transport),
		},
	}
}

// CertResponse contains the response from Vault's cert endpoint.
type CertResponse struct {
	Data CertData `json:"data"`
}

// CertData contains the certificate data from Vault.
type CertData struct {
	Certificate  string `json:"certificate"`
	RevocationTime    int64  `json:"revocation_time"`
	RevocationTimeRFC string `json:"revocation_time_rfc3339"`
}

// listResponse is the Vault LIST response for certificates.
type listResponse struct {
	Data struct {
		Keys []string `json:"keys"`
	} `json:"data"`
}

// ListCertSerials returns all certificate serial numbers from a PKI mount.
// Uses LIST /v1/{mount}/certs.
func (c *Client) ListCertSerials(ctx context.Context, mountPath string) ([]string, error) {
	url := fmt.Sprintf("%s/v1/%scerts", c.baseURL, ensureTrailingSlash(mountPath))

	req, err := http.NewRequestWithContext(ctx, "LIST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list certs: %w", err)
	}
	defer resp.Body.Close()

	// 404 means no certs exist
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list certs: status %d: %s", resp.StatusCode, string(body))
	}

	var listResp listResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("decode list response: %w", err)
	}

	return listResp.Data.Keys, nil
}

// GetCert returns the certificate for a given serial number.
// Uses GET /v1/{mount}/cert/{serial}.
func (c *Client) GetCert(ctx context.Context, mountPath, serial string) (*CertResponse, error) {
	url := fmt.Sprintf("%s/v1/%scert/%s", c.baseURL, ensureTrailingSlash(mountPath), serial)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get cert %s: %w", serial, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get cert %s: status %d: %s", serial, resp.StatusCode, string(body))
	}

	var certResp CertResponse
	if err := json.NewDecoder(resp.Body).Decode(&certResp); err != nil {
		return nil, fmt.Errorf("decode cert response: %w", err)
	}

	return &certResp, nil
}

func ensureTrailingSlash(s string) string {
	if !strings.HasSuffix(s, "/") {
		return s + "/"
	}
	return s
}
