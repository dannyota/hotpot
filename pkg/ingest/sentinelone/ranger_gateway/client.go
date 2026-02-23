package ranger_gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client wraps the SentinelOne Ranger Gateways API.
type Client struct {
	baseURL    string
	apiToken   string
	batchSize  int
	httpClient *http.Client
}

// NewClient creates a new SentinelOne ranger gateways client.
func NewClient(baseURL, apiToken string, batchSize int, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		batchSize:  batchSize,
		httpClient: httpClient,
	}
}

// APIRangerGateway represents the gateway data from the SentinelOne Ranger API response.
type APIRangerGateway struct {
	ID                   string          `json:"id"`
	IP                   string          `json:"ip"`
	MacAddress           string          `json:"macAddress"`
	ExternalIP           string          `json:"externalIp"`
	Manufacturer         string          `json:"manufacturer"`
	NetworkName          string          `json:"networkName"`
	AccountID            json.Number     `json:"accountId"`
	AccountName          string          `json:"accountName"`
	SiteID               json.Number     `json:"siteId"`
	NumberOfAgents       int             `json:"numberOfAgents"`
	NumberOfRangers      int             `json:"numberOfRangers"`
	ConnectedRangers     int             `json:"connectedRangers"`
	TotalAgents          int             `json:"totalAgents"`
	AgentPercentage      float64         `json:"agentPercentage"`
	AllowScan            bool            `json:"allowScan"`
	Archived             bool            `json:"archived"`
	New                  bool            `json:"new"`
	InheritSettings      bool            `json:"inheritSettings"`
	TCPPortScan          bool            `json:"tcpPortScan"`
	UDPPortScan          bool            `json:"udpPortScan"`
	ICMPScan             bool            `json:"icmpScan"`
	SMBScan              bool            `json:"smbScan"`
	MDNSScan             bool            `json:"mdnsScan"`
	RDNSScan             bool            `json:"rdnsScan"`
	SNMPScan             bool            `json:"snmpScan"`
	ScanOnlyLocalSubnets bool            `json:"scanOnlyLocalSubnets"`
	CreatedAt            *time.Time      `json:"createdAt"`
	ExpiryDate           *time.Time      `json:"expiryDate"`
	TCPPorts             json.RawMessage `json:"tcpPorts"`
	UDPPorts             json.RawMessage `json:"udpPorts"`
	Restrictions         json.RawMessage `json:"restrictions"`
}

// GatewayBatchResult contains a batch of gateways and pagination info.
type GatewayBatchResult struct {
	Gateways   []APIRangerGateway
	NextCursor string
	HasMore    bool
}

// GetGatewaysBatch retrieves a batch of ranger gateways with cursor pagination.
func (c *Client) GetGatewaysBatch(cursor string) (*GatewayBatchResult, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", c.batchSize))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.doRequest("GET", "/web/api/v2.0/ranger/gateways", params)
	if err != nil {
		return nil, fmt.Errorf("get ranger gateways: %w", err)
	}

	var response struct {
		Data       []APIRangerGateway `json:"data"`
		Pagination struct {
			NextCursor string `json:"nextCursor"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse ranger gateways response: %w", err)
	}

	return &GatewayBatchResult{
		Gateways:   response.Data,
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
