package config

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// VaultSource reads config from HashiCorp Vault KV v2.
type VaultSource struct {
	address      string
	token        string
	secretPath   string
	mount        string
	httpClient   *http.Client
	pollInterval time.Duration
}

// VaultSourceOptions configures VaultSource.
type VaultSourceOptions struct {
	// Address is the Vault server address (e.g., https://vault.example.com).
	Address string

	// Token is the Vault authentication token.
	Token string

	// SecretPath is the path to the secret (e.g., "hotpot/config").
	SecretPath string

	// Mount is the KV v2 mount point (default: "secret").
	Mount string

	// VerifySSL enables TLS certificate verification.
	VerifySSL bool

	// Timeout is the HTTP request timeout (default: 30s).
	Timeout time.Duration

	// PollInterval is how often to poll for changes (default: 30s).
	PollInterval time.Duration
}

// NewVaultSource creates a Vault-based config source.
func NewVaultSource(opts VaultSourceOptions) *VaultSource {
	if opts.Mount == "" {
		opts.Mount = "secret"
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.PollInterval == 0 {
		opts.PollInterval = 30 * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !opts.VerifySSL,
		},
	}

	return &VaultSource{
		address:      opts.Address,
		token:        opts.Token,
		secretPath:   opts.SecretPath,
		mount:        opts.Mount,
		pollInterval: opts.PollInterval,
		httpClient: &http.Client{
			Timeout:   opts.Timeout,
			Transport: transport,
		},
	}
}

// Load reads config from Vault KV v2.
func (v *VaultSource) Load(ctx context.Context) (*Config, error) {
	secret, err := v.readSecret(ctx)
	if err != nil {
		return nil, err
	}

	return v.parseConfig(secret)
}

// Watch polls Vault for changes using secret version metadata.
func (v *VaultSource) Watch(ctx context.Context, onChange func()) (func(), error) {
	done := make(chan struct{})

	// Track last known version
	var lastVersion int64 = -1

	go func() {
		ticker := time.NewTicker(v.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case <-ticker.C:
				version, err := v.getSecretVersion(ctx)
				if err != nil {
					fmt.Printf("vault version check error: %v\n", err)
					continue
				}
				if lastVersion >= 0 && version != lastVersion {
					onChange()
				}
				lastVersion = version
			}
		}
	}()

	return func() { close(done) }, nil
}

// Type returns "vault".
func (v *VaultSource) Type() string {
	return "vault"
}

// readSecret reads the secret data from Vault KV v2.
func (v *VaultSource) readSecret(ctx context.Context) (map[string]interface{}, error) {
	// KV v2 API: /v1/{mount}/data/{path}
	url := fmt.Sprintf("%s/v1/%s/data/%s", v.address, v.mount, v.secretPath)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", v.token)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vault request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("secret not found at path: %s", v.secretPath)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vault API status %d: %s", resp.StatusCode, string(body))
	}

	var secretResp struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&secretResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return secretResp.Data.Data, nil
}

// getSecretVersion returns the current version of the secret.
func (v *VaultSource) getSecretVersion(ctx context.Context) (int64, error) {
	// KV v2 metadata API: /v1/{mount}/metadata/{path}
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", v.address, v.mount, v.secretPath)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", v.token)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("vault request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("vault metadata API status %d", resp.StatusCode)
	}

	var metaResp struct {
		Data struct {
			CurrentVersion int64 `json:"current_version"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&metaResp); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	return metaResp.Data.CurrentVersion, nil
}

// parseConfig converts Vault secret data to Config struct.
func (v *VaultSource) parseConfig(data map[string]interface{}) (*Config, error) {
	cfg := &Config{}

	// GCP credentials
	if val, ok := data["gcp_credentials_json"].(string); ok && val != "" {
		cfg.GCP.CredentialsJSON = []byte(val)
	}
	// Database config
	if val, ok := data["database_host"].(string); ok {
		cfg.Database.Host = val
	}
	if val, ok := data["database_port"]; ok {
		cfg.Database.Port = toInt(val)
	}
	if val, ok := data["database_user"].(string); ok {
		cfg.Database.User = val
	}
	if val, ok := data["database_password"].(string); ok {
		cfg.Database.Password = val
	}
	if val, ok := data["database_dbname"].(string); ok {
		cfg.Database.DBName = val
	}
	if val, ok := data["database_sslmode"].(string); ok {
		cfg.Database.SSLMode = val
	}

	// Temporal config
	if val, ok := data["temporal_host_port"].(string); ok {
		cfg.Temporal.HostPort = val
	}
	if val, ok := data["temporal_namespace"].(string); ok {
		cfg.Temporal.Namespace = val
	}

	return cfg, nil
}

// toInt converts interface{} to int (handles both float64 from JSON and string).
func toInt(val interface{}) int {
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		i, _ := strconv.Atoi(v)
		return i
	default:
		return 0
	}
}
