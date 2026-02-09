package account

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client wraps the SentinelOne Accounts API.
type Client struct {
	baseURL    string
	apiToken   string
	batchSize  int
	httpClient *http.Client
}

// NewClient creates a new SentinelOne accounts client.
func NewClient(baseURL, apiToken string, batchSize int, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		batchSize:  batchSize,
		httpClient: httpClient,
	}
}

// APIAccount represents the account data from the SentinelOne API response.
type APIAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetAccounts retrieves all accounts.
func (c *Client) GetAccounts() ([]APIAccount, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", c.batchSize))

	body, err := c.doRequest("GET", "/web/api/v2.1/accounts", params)
	if err != nil {
		return nil, fmt.Errorf("get accounts: %w", err)
	}

	var response struct {
		Data []APIAccount `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse accounts response: %w", err)
	}

	return response.Data, nil
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
