package site

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client wraps the SentinelOne Sites API.
type Client struct {
	baseURL    string
	apiToken   string
	batchSize  int
	httpClient *http.Client
}

// NewClient creates a new SentinelOne sites client.
func NewClient(baseURL, apiToken string, batchSize int, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		batchSize:  batchSize,
		httpClient: httpClient,
	}
}

// APISite represents the site data from the SentinelOne API response.
type APISite struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	AccountID         string  `json:"accountId"`
	AccountName       string  `json:"accountName"`
	State             string  `json:"state"`
	SiteType          string  `json:"siteType"`
	Suite             string  `json:"suite"`
	Creator           string  `json:"creator"`
	CreatorID         string  `json:"creatorId"`
	HealthStatus      bool    `json:"healthStatus"`
	ActiveLicenses    int     `json:"activeLicenses"`
	TotalLicenses     int     `json:"totalLicenses"`
	UnlimitedLicenses bool    `json:"unlimitedLicenses"`
	IsDefault         bool    `json:"isDefault"`
	Description       string  `json:"description"`
	CreatedAt         *string `json:"createdAt"`
	Expiration               *string         `json:"expiration"`
	UpdatedAt                *string         `json:"updatedAt"`
	ExternalID               string          `json:"externalId"`
	SKU                      string          `json:"sku"`
	UsageType                string          `json:"usageType"`
	UnlimitedExpiration      bool            `json:"unlimitedExpiration"`
	InheritAccountExpiration bool            `json:"inheritAccountExpiration"`
	Licenses                 json.RawMessage `json:"licenses"`
}

// SiteBatchResult contains a batch of sites and pagination info.
type SiteBatchResult struct {
	Sites      []APISite
	NextCursor string
	HasMore    bool
}

// GetSitesBatch retrieves a batch of sites with cursor pagination.
// Note: The SentinelOne Sites API nests data under data.sites[] instead of data[].
func (c *Client) GetSitesBatch(cursor string) (*SiteBatchResult, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", c.batchSize))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.doRequest("GET", "/web/api/v2.1/sites", params)
	if err != nil {
		return nil, fmt.Errorf("get sites: %w", err)
	}

	var response struct {
		Data struct {
			Sites []APISite `json:"sites"`
		} `json:"data"`
		Pagination struct {
			NextCursor string `json:"nextCursor"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse sites response: %w", err)
	}

	return &SiteBatchResult{
		Sites:      response.Data.Sites,
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
