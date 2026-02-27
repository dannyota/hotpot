package endpoint_app

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/dannyota/hotpot/pkg/base/httperr"
)

// Client wraps the SentinelOne Endpoint Applications API.
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new SentinelOne endpoint apps client.
func NewClient(baseURL, apiToken string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		httpClient: httpClient,
	}
}

// APIEndpointApp represents an application installed on an agent from the SentinelOne API.
type APIEndpointApp struct {
	Name          string  `json:"name"`
	Version       string  `json:"version"`
	Publisher     string  `json:"publisher"`
	Size          int     `json:"size"`
	InstalledDate *string `json:"installedDate"`
}

// GetEndpointApps retrieves all applications for a single agent.
func (c *Client) GetEndpointApps(agentID string) ([]APIEndpointApp, error) {
	params := url.Values{}
	params.Set("ids", agentID)

	body, err := c.doRequest("GET", "/web/api/v2.1/inventory/applications", params)
	if err != nil {
		return nil, fmt.Errorf("get endpoint apps for agent %s: %w", agentID, err)
	}

	var response struct {
		Data []APIEndpointApp `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse endpoint apps response: %w", err)
	}

	return response.Data, nil
}

func (c *Client) doRequest(method, endpoint string, params url.Values) ([]byte, error) {
	requestURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	if params != nil {
		requestURL = fmt.Sprintf("%s?%s", requestURL, params.Encode())
	}
	start := time.Now()

	req, err := http.NewRequest(method, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("ApiToken %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")

	slog.Debug("s1 api request", "method", method, "endpoint", endpoint)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("s1 api request failed", "method", method, "endpoint", endpoint, "error", err, "durationMs", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	slog.Info("s1 api response", "method", method, "endpoint", endpoint, "status", resp.StatusCode, "responseBytes", len(body), "durationMs", time.Since(start).Milliseconds())

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, &httperr.APIError{Code: resp.StatusCode}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
