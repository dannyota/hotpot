package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client wraps the SentinelOne Installed Applications API.
type Client struct {
	baseURL    string
	apiToken   string
	batchSize  int
	httpClient *http.Client
}

// NewClient creates a new SentinelOne installed applications client.
func NewClient(baseURL, apiToken string, batchSize int, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		batchSize:  batchSize,
		httpClient: httpClient,
	}
}

// APIApp represents the installed application data from the SentinelOne API response.
type APIApp struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	Publisher             string  `json:"publisher"`
	Version               string  `json:"version"`
	Size                  int64   `json:"size"`
	Type                  string  `json:"type"`
	OsType                string  `json:"osType"`
	InstalledDate         *string `json:"installedAt"`
	AgentID               string  `json:"agentId"`
	AgentComputerName     string  `json:"agentComputerName"`
	AgentMachineType      string  `json:"agentMachineType"`
	AgentIsActive         bool    `json:"agentIsActive"`
	AgentIsDecommissioned bool    `json:"agentIsDecommissioned"`
	RiskLevel             string  `json:"riskLevel"`
	Signed                bool    `json:"signed"`
	CreatedAt             *string `json:"createdAt"`
	UpdatedAt             *string `json:"updatedAt"`
	AgentUUID             string  `json:"agentUuid"`
	AgentDomain           string  `json:"agentDomain"`
	AgentVersion          string  `json:"agentVersion"`
	AgentOsType           string  `json:"agentOsType"`
	AgentNetworkStatus    string  `json:"agentNetworkStatus"`
	AgentInfected         bool    `json:"agentInfected"`
	AgentOperationalState string  `json:"agentOperationalState"`
}

// AppBatchResult contains a batch of apps and pagination info.
type AppBatchResult struct {
	Apps       []APIApp
	NextCursor string
	HasMore    bool
}

// GetAppsBatch retrieves a batch of installed applications with cursor pagination.
func (c *Client) GetAppsBatch(cursor string) (*AppBatchResult, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", c.batchSize))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.doRequest("GET", "/web/api/v2.1/installed-applications", params)
	if err != nil {
		return nil, fmt.Errorf("get installed applications: %w", err)
	}

	var response struct {
		Data       []APIApp `json:"data"`
		Pagination struct {
			NextCursor string `json:"nextCursor"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse installed applications response: %w", err)
	}

	return &AppBatchResult{
		Apps:       response.Data,
		NextCursor: response.Pagination.NextCursor,
		HasMore:    response.Pagination.NextCursor != "",
	}, nil
}

func (c *Client) doRequest(method, endpoint string, params url.Values) ([]byte, error) {
	requestURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	if params != nil {
		requestURL = fmt.Sprintf("%s?%s", requestURL, params.Encode())
	}

	req, err := http.NewRequest(method, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("ApiToken %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("authentication failed (status: %d)", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
