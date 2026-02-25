package ranger_device

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// Client wraps the SentinelOne Ranger API.
type Client struct {
	baseURL    string
	apiToken   string
	batchSize  int
	httpClient *http.Client
}

// NewClient creates a new SentinelOne ranger device client.
func NewClient(baseURL, apiToken string, batchSize int, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		batchSize:  batchSize,
		httpClient: httpClient,
	}
}

// APIRangerDevice represents the ranger device data from the SentinelOne API response.
type APIRangerDevice struct {
	ID                string          `json:"id"`
	LocalIP           string          `json:"localIp"`
	ExternalIP        string          `json:"externalIp"`
	MacAddress        string          `json:"macAddress"`
	OsType            string          `json:"osType"`
	OsName            string          `json:"osName"`
	OsVersion         string          `json:"osVersion"`
	DeviceType        string          `json:"deviceType"`
	DeviceFunction    string          `json:"deviceFunction"`
	Manufacturer      string          `json:"manufacturer"`
	ManagedState      string          `json:"managedState"`
	AgentID           string          `json:"agentId"`
	FirstSeen         *time.Time      `json:"firstSeen"`
	LastSeen          *time.Time      `json:"lastSeen"`
	SubnetAddress     string          `json:"subnetAddress"`
	GatewayIPAddress  string          `json:"gatewayIpAddress"`
	GatewayMacAddress string          `json:"gatewayMacAddress"`
	NetworkName       string          `json:"networkName"`
	Domain            string          `json:"domain"`
	SiteName          string          `json:"siteName"`
	DeviceReview      string          `json:"deviceReview"`
	HasIdentity       bool            `json:"hasIdentity"`
	HasUserLabel      bool            `json:"hasUserLabel"`
	FingerprintScore  int             `json:"fingerprintScore"`
	TCPPorts          json.RawMessage `json:"tcpPorts"`
	UDPPorts          json.RawMessage `json:"udpPorts"`
	Hostnames         json.RawMessage `json:"hostnames"`
	DiscoveryMethods  json.RawMessage `json:"discoveryMethods"`
	Networks          json.RawMessage `json:"networks"`
	Tags              json.RawMessage `json:"tags"`
}

// DeviceBatchResult contains a batch of ranger devices and pagination info.
type DeviceBatchResult struct {
	Devices    []APIRangerDevice
	NextCursor string
	HasMore    bool
}

// GetDevicesBatch retrieves a batch of ranger devices with cursor pagination.
func (c *Client) GetDevicesBatch(cursor string) (*DeviceBatchResult, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", c.batchSize))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.doRequest("GET", "/web/api/v2.0/ranger/table-view", params)
	if err != nil {
		return nil, fmt.Errorf("get ranger devices: %w", err)
	}

	var response struct {
		Data       []APIRangerDevice `json:"data"`
		Pagination struct {
			NextCursor string `json:"nextCursor"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse ranger devices response: %w", err)
	}

	return &DeviceBatchResult{
		Devices:    response.Data,
		NextCursor: response.Pagination.NextCursor,
		HasMore:    response.Pagination.NextCursor != "",
	}, nil
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
		return nil, fmt.Errorf("authentication failed (status: %d)", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
