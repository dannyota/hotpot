package ranger_setting

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

// Client wraps the SentinelOne Ranger Settings API.
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new SentinelOne Ranger settings client.
func NewClient(baseURL, apiToken string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		httpClient: httpClient,
	}
}

// APIRangerSetting represents the settings data from the SentinelOne Ranger API.
type APIRangerSetting struct {
	AccountID                json.Number     `json:"accountId"`
	ScopeID                  string          `json:"scopeId"`
	Enabled                  bool            `json:"enabled"`
	UsePeriodicSnapshots     bool            `json:"usePeriodicSnapshots"`
	SnapshotPeriod           int             `json:"snapshotPeriod"`
	NetworkDecommissionValue int             `json:"networkDecommissionValue"`
	MinAgentsInNetworkToScan int             `json:"minAgentsInNetworkToScan"`
	TCPPortScan              bool            `json:"tcpPortScan"`
	UDPPortScan              bool            `json:"udpPortScan"`
	ICMPScan                 bool            `json:"icmpScan"`
	SMBScan                  bool            `json:"smbScan"`
	MDNSScan                 bool            `json:"mdnsScan"`
	RDNSScan                 bool            `json:"rdnsScan"`
	SNMPScan                 bool            `json:"snmpScan"`
	MultiScanSSDP            bool            `json:"multiScanSsdp"`
	UseFullDNSScan           bool            `json:"useFullDnsScan"`
	ScanOnlyLocalSubnets     bool            `json:"scanOnlyLocalSubnets"`
	AutoEnableNetworks       bool            `json:"autoEnableNetworks"`
	CombineDevices           bool            `json:"combineDevices"`
	NewNetworkInHours        int             `json:"newNetworkInHours"`
	TCPPorts                 json.RawMessage `json:"tcpPorts"`
	UDPPorts                 json.RawMessage `json:"udpPorts"`
	Restrictions             json.RawMessage `json:"restrictions"`
}

// GetSettings retrieves Ranger settings for a specific account.
func (c *Client) GetSettings(accountID string) (*APIRangerSetting, error) {
	params := url.Values{}
	params.Set("accountIds", accountID)

	body, err := c.doRequest("GET", "/web/api/v2.0/ranger/settings", params)
	if err != nil {
		return nil, fmt.Errorf("get ranger settings: %w", err)
	}

	var response struct {
		Data APIRangerSetting `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse ranger settings response: %w", err)
	}

	return &response.Data, nil
}

func (c *Client) doRequest(method, endpoint string, params url.Values) ([]byte, error) {
	requestURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	if params != nil {
		requestURL = fmt.Sprintf("%s?%s", requestURL, params.Encode())
	}

	start := time.Now()
	slog.Debug("s1 api request", "method", method, "endpoint", endpoint)

	req, err := http.NewRequest(method, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("ApiToken %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")

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
