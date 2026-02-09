package threat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client wraps the SentinelOne Threats API.
type Client struct {
	baseURL    string
	apiToken   string
	batchSize  int
	httpClient *http.Client
}

// NewClient creates a new SentinelOne threats client.
func NewClient(baseURL, apiToken string, batchSize int, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		batchSize:  batchSize,
		httpClient: httpClient,
	}
}

// APIThreat represents the threat data from the SentinelOne API response.
type APIThreat struct {
	ID              string          `json:"id"`
	AgentID         string          `json:"agentRealtimeInfo.agentId,omitempty"`
	Classification  string          `json:"classification"`
	ThreatName      string          `json:"threatName"`
	FilePath        string          `json:"filePath"`
	MitigationStatus string        `json:"mitigationStatus"`
	AnalystVerdict  string          `json:"analystVerdict"`
	ConfidenceLevel string          `json:"confidenceLevel"`
	InitiatedBy     string          `json:"initiatedBy"`
	CreatedAt       *time.Time      `json:"createdDate"`
	ThreatInfo      json.RawMessage `json:"threatInfo"`

	UpdatedAt            *time.Time `json:"updatedAt"`
	FileContentHash      string     `json:"fileContentHash"`
	CloudVerdict         string     `json:"cloudVerdict"`
	ClassificationSource string     `json:"classificationSource"`

	// The API nests agent info under agentRealtimeInfo
	AgentRealtimeInfo struct {
		AgentID              string `json:"agentId"`
		SiteID               string `json:"siteId"`
		SiteName             string `json:"siteName"`
		AccountID            string `json:"accountId"`
		AccountName          string `json:"accountName"`
		AgentComputerName    string `json:"agentComputerName"`
		AgentOsType          string `json:"agentOsType"`
		AgentMachineType     string `json:"agentMachineType"`
		AgentIsActive        bool   `json:"agentIsActive"`
		AgentIsDecommissioned bool  `json:"agentIsDecommissioned"`
		AgentVersion         string `json:"agentVersion"`
	} `json:"agentRealtimeInfo"`
}

// ThreatInfoData is used to extract specific fields from the threatInfo JSON blob.
type ThreatInfoData struct {
	SHA256 string `json:"sha256"`
}

// ThreatBatchResult contains a batch of threats and pagination info.
type ThreatBatchResult struct {
	Threats    []APIThreat
	NextCursor string
	HasMore    bool
}

// GetThreatsBatch retrieves a batch of threats with cursor pagination.
func (c *Client) GetThreatsBatch(cursor string) (*ThreatBatchResult, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", c.batchSize))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.doRequest("GET", "/web/api/v2.1/threats", params)
	if err != nil {
		return nil, fmt.Errorf("get threats: %w", err)
	}

	var response struct {
		Data       []APIThreat `json:"data"`
		Pagination struct {
			NextCursor string `json:"nextCursor"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse threats response: %w", err)
	}

	return &ThreatBatchResult{
		Threats:    response.Data,
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
