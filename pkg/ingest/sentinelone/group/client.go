package group

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client wraps the SentinelOne Groups API.
type Client struct {
	baseURL    string
	apiToken   string
	batchSize  int
	httpClient *http.Client
}

// NewClient creates a new SentinelOne groups client.
func NewClient(baseURL, apiToken string, batchSize int, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		batchSize:  batchSize,
		httpClient: httpClient,
	}
}

// APIGroup represents the group data from the SentinelOne API response.
type APIGroup struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	SiteID      string  `json:"siteId"`
	Type        string  `json:"type"`
	IsDefault   bool    `json:"isDefault"`
	Inherits    bool    `json:"inherits"`
	Rank        *int    `json:"rank"`
	TotalAgents int     `json:"totalAgents"`
	Creator     string  `json:"creator"`
	CreatorID   string  `json:"creatorId"`
	FilterName  string  `json:"filterName"`
	FilterID    string  `json:"filterId"`
	CreatedAt         *string `json:"createdAt"`
	UpdatedAt         *string `json:"updatedAt"`
	RegistrationToken string  `json:"registrationToken"`
}

// GroupBatchResult contains a batch of groups and pagination info.
type GroupBatchResult struct {
	Groups     []APIGroup
	NextCursor string
	HasMore    bool
}

// GetGroupsBatch retrieves a batch of groups with cursor pagination.
func (c *Client) GetGroupsBatch(cursor string) (*GroupBatchResult, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", c.batchSize))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.doRequest("GET", "/web/api/v2.1/groups", params)
	if err != nil {
		return nil, fmt.Errorf("get groups: %w", err)
	}

	var response struct {
		Data       []APIGroup `json:"data"`
		Pagination struct {
			NextCursor string `json:"nextCursor"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse groups response: %w", err)
	}

	return &GroupBatchResult{
		Groups:     response.Data,
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
