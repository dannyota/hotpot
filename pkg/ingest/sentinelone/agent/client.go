package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client wraps the SentinelOne Agents API.
type Client struct {
	baseURL    string
	apiToken   string
	batchSize  int
	httpClient *http.Client
}

// NewClient creates a new SentinelOne agents client.
func NewClient(baseURL, apiToken string, batchSize int, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		batchSize:  batchSize,
		httpClient: httpClient,
	}
}

// APINetworkInterface represents a network interface from the SentinelOne API response.
type APINetworkInterface struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Inet        []string `json:"inet"`
	Inet6       []string `json:"inet6"`
	Physical    string   `json:"physical"`
	GatewayIP   string   `json:"gatewayIp"`
	GatewayMac  string   `json:"gatewayMacAddress"`
}

// APIAgent represents the agent data from the SentinelOne API response.
type APIAgent struct {
	ID                      string                `json:"id"`
	ComputerName            string                `json:"computerName"`
	ExternalIP              string                `json:"externalIp"`
	SiteName                string                `json:"siteName"`
	AccountID               string                `json:"accountId"`
	AccountName             string                `json:"accountName"`
	AgentVersion            string                `json:"agentVersion"`
	OSType                  string                `json:"osType"`
	OSName                  string                `json:"osName"`
	OSRevision              string                `json:"osRevision"`
	OSArch                  string                `json:"osArch"`
	IsActive                bool                  `json:"isActive"`
	IsInfected              bool                  `json:"infected"`
	IsDecommissioned        bool                  `json:"isDecommissioned"`
	MachineType             string                `json:"machineType"`
	Domain                  string                `json:"domain"`
	UUID                    string                `json:"uuid"`
	NetworkStatus           string                `json:"networkStatus"`
	LastActiveDate          *time.Time            `json:"lastActiveDate"`
	RegisteredAt            *time.Time            `json:"registeredAt"`
	UpdatedAt               *time.Time            `json:"updatedAt"`
	OSStartTime             *time.Time            `json:"osStartTime"`
	ActiveThreats           int                   `json:"activeThreats"`
	EncryptedApplications   bool                  `json:"encryptedApplications"`
	GroupName               string                `json:"groupName"`
	GroupID                 string                `json:"groupId"`
	CPUCount                int                   `json:"cpuCount"`
	CoreCount               int                   `json:"coreCount"`
	CPUId                   string                `json:"cpuId"`
	TotalMemory             int64                 `json:"totalMemory"`
	ModelName               string                `json:"modelName"`
	SerialNumber            string                `json:"serialNumber"`
	StorageEncryptionStatus string                `json:"storageEncryptionStatus"`
	NetworkInterfaces       []APINetworkInterface `json:"networkInterfaces"`
}

// AgentBatchResult contains a batch of agents and pagination info.
type AgentBatchResult struct {
	Agents     []APIAgent
	NextCursor string
	HasMore    bool
}

// GetAgentsBatch retrieves a batch of agents with cursor pagination.
func (c *Client) GetAgentsBatch(cursor string) (*AgentBatchResult, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", c.batchSize))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.doRequest("GET", "/web/api/v2.1/agents", params)
	if err != nil {
		return nil, fmt.Errorf("get agents: %w", err)
	}

	var response struct {
		Data       []APIAgent `json:"data"`
		Pagination struct {
			NextCursor string `json:"nextCursor"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse agents response: %w", err)
	}

	return &AgentBatchResult{
		Agents:     response.Data,
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
