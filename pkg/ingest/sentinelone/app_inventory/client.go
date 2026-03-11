package app_inventory

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"danny.vn/hotpot/pkg/base/httperr"
)

// Client wraps the SentinelOne Application Inventory API.
type Client struct {
	baseURL    string
	apiToken   string
	batchSize  int
	httpClient *http.Client
}

// NewClient creates a new SentinelOne app inventory client.
func NewClient(baseURL, apiToken string, batchSize int, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		batchSize:  batchSize,
		httpClient: httpClient,
	}
}

// APIAppInventory represents an application from the SentinelOne /inventory API.
type APIAppInventory struct {
	ApplicationName          string `json:"applicationName"`
	ApplicationVendor        string `json:"applicationVendor"`
	EndpointsCount           int    `json:"endpointsCount"`
	ApplicationVersionsCount int    `json:"applicationVersionsCount"`
	Estimate                 bool   `json:"estimate"`
}

// AppBatchResult contains a batch of app inventory entries and pagination info.
type AppBatchResult struct {
	Apps       []APIAppInventory
	NextCursor string
	HasMore    bool
	TotalItems int
}

// GetAppsBatch retrieves a batch of application inventory entries with cursor pagination.
func (c *Client) GetAppsBatch(cursor string) (*AppBatchResult, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", c.batchSize))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.doRequest("GET", "/web/api/v2.1/application-management/inventory", params)
	if err != nil {
		return nil, fmt.Errorf("get app inventory: %w", err)
	}

	var response struct {
		Data       []APIAppInventory `json:"data"`
		Pagination struct {
			NextCursor string `json:"nextCursor"`
			TotalItems int    `json:"totalItems"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse app inventory response: %w", err)
	}

	return &AppBatchResult{
		Apps:       response.Data,
		NextCursor: response.Pagination.NextCursor,
		HasMore:    response.Pagination.NextCursor != "",
		TotalItems: response.Pagination.TotalItems,
	}, nil
}

// GetCount returns the total number of app inventory entries using countOnly mode.
func (c *Client) GetCount() (int, error) {
	params := url.Values{}
	params.Set("countOnly", "true")

	body, err := c.doRequest("GET", "/web/api/v2.1/application-management/inventory", params)
	if err != nil {
		return 0, fmt.Errorf("get app inventory count: %w", err)
	}

	var response struct {
		Pagination struct {
			TotalItems int `json:"totalItems"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("parse app inventory count response: %w", err)
	}

	return response.Pagination.TotalItems, nil
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
