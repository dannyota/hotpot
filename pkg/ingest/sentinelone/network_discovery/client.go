package network_discovery

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

// Client wraps the SentinelOne XDR network discovery API.
type Client struct {
	baseURL    string
	apiToken   string
	batchSize  int
	httpClient *http.Client
}

// NewClient creates a new SentinelOne network discovery client.
func NewClient(baseURL, apiToken string, batchSize int, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		batchSize:  batchSize,
		httpClient: httpClient,
	}
}

// APINetworkDiscovery represents the network discovery device data from the SentinelOne API response.
type APINetworkDiscovery struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	IPAddress    string `json:"ipAddress"`
	Domain       string `json:"domain"`
	SerialNumber string `json:"serialNumber"`
	Category     string `json:"category"`
	SubCategory  string `json:"subCategory"`
	ResourceType string `json:"resourceType"`
	OS           string `json:"os"`
	OSFamily     string `json:"osFamily"`
	OSVersion    string `json:"osVersion"`
	OSNameVersion string `json:"osNameVersion"`
	Architecture string `json:"architecture"`
	Manufacturer string `json:"manufacturer"`
	CPU          string `json:"cpu"`
	MemoryReadable string `json:"memoryReadable"`
	NetworkName    string `json:"networkName"`
	AssetStatus    string `json:"assetStatus"`
	AssetCriticality   string `json:"assetCriticality"`
	AssetEnvironment   string `json:"assetEnvironment"`
	InfectionStatus    string `json:"infectionStatus"`
	DeviceReview       string `json:"deviceReview"`
	EppUnsupportedUnknown    string `json:"eppUnsupportedUnknown"`
	AssetContactEmail        string `json:"assetContactEmail"`
	LegacyIdentityPolicyName string `json:"legacyIdentityPolicyName"`
	PreviousOSType           string `json:"previousOsType"`
	PreviousOSVersion        string `json:"previousOsVersion"`
	PreviousDeviceFunction   string `json:"previousDeviceFunction"`
	DetectedFromSite         string `json:"detectedFromSite"`
	S1AccountID    string `json:"s1AccountId"`
	S1AccountName  string `json:"s1AccountName"`
	S1SiteID       string `json:"s1SiteId"`
	S1SiteName     string `json:"s1SiteName"`
	S1GroupID      string `json:"s1GroupId"`
	S1GroupName    string `json:"s1GroupName"`
	S1ScopeID      string `json:"s1ScopeId"`
	S1ScopeLevel   string `json:"s1ScopeLevel"`
	S1ScopePath    string `json:"s1ScopePath"`
	S1OnboardedAccountName string `json:"s1OnboardedAccountName"`
	S1OnboardedGroupName   string `json:"s1OnboardedGroupName"`
	S1OnboardedSiteName    string `json:"s1OnboardedSiteName"`
	S1OnboardedScopeLevel  string `json:"s1OnboardedScopeLevel"`
	S1OnboardedScopePath   string `json:"s1OnboardedScopePath"`

	// Int fields
	Memory              int `json:"memory"`
	CoreCount           int `json:"coreCount"`
	S1ManagementID      int `json:"s1ManagementId"`
	S1ScopeType         int `json:"s1ScopeType"`
	S1OnboardedAccountID int `json:"s1OnboardedAccountId"`
	S1OnboardedGroupID   int `json:"s1OnboardedGroupId"`
	S1OnboardedScopeID   int `json:"s1OnboardedScopeId"`
	S1OnboardedSiteID    int `json:"s1OnboardedSiteId"`

	// Bool fields
	IsAdConnector bool `json:"isAdConnector"`
	IsDcServer    bool `json:"isDcServer"`
	AdsEnabled    bool `json:"adsEnabled"`

	// Time fields
	FirstSeenDt  *time.Time `json:"firstSeenDt"`
	LastUpdateDt *time.Time `json:"lastUpdateDt"`
	LastActiveDt *time.Time `json:"lastActiveDt"`
	LastRebootDt *time.Time `json:"lastRebootDt"`
	S1UpdatedAt  *time.Time `json:"s1UpdatedAt"`

	// JSON fields
	AgentJSON             json.RawMessage `json:"agent"`
	NetworkInterfacesJSON json.RawMessage `json:"networkInterfaces"`
	AlertsJSON            json.RawMessage `json:"alerts"`
	AlertsCountJSON       json.RawMessage `json:"alertsCount"`
	DeviceReviewLogJSON   json.RawMessage `json:"deviceReviewLog"`
	IdentityJSON          json.RawMessage `json:"identity"`
	NotesJSON             json.RawMessage `json:"notes"`
	TagsJSON              json.RawMessage `json:"tags"`
	MissingCoverageJSON   json.RawMessage `json:"missingCoverage"`
	SubnetsJSON           json.RawMessage `json:"subnets"`
	SurfacesJSON          json.RawMessage `json:"surfaces"`
	NetworkNamesJSON      json.RawMessage `json:"networkNames"`
	RiskFactorsJSON       json.RawMessage `json:"riskFactors"`
	ActiveCoverageJSON    json.RawMessage `json:"activeCoverage"`
	DiscoveryMethodsJSON  json.RawMessage `json:"discoveryMethods"`
	HostnamesJSON         json.RawMessage `json:"hostnames"`
	InternalIPsJSON       json.RawMessage `json:"internalIps"`
	InternalIPsV6JSON     json.RawMessage `json:"internalIpsV6"`
	MACAddressesJSON      json.RawMessage `json:"macAddresses"`
	GatewayIPsJSON        json.RawMessage `json:"gatewayIps"`
	GatewayMacsJSON       json.RawMessage `json:"gatewayMacs"`
	TCPPortsJSON          json.RawMessage `json:"tcpPorts"`
	UDPPortsJSON          json.RawMessage `json:"udpPorts"`
	RangerTagsJSON        json.RawMessage `json:"rangerTags"`
	IDSecondaryJSON       json.RawMessage `json:"idSecondary"`
}

// NetworkDiscoveryBatchResult contains a batch of network discovery devices and pagination info.
type NetworkDiscoveryBatchResult struct {
	Devices    []APINetworkDiscovery
	NextCursor string
	HasMore    bool
}

// GetDevicesBatch retrieves a batch of network discovery devices with cursor pagination.
func (c *Client) GetDevicesBatch(cursor string) (*NetworkDiscoveryBatchResult, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", c.batchSize))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.doRequest("GET", "/web/api/v2.1/xdr/assets/surface/networkDiscovery", params)
	if err != nil {
		return nil, fmt.Errorf("get network discovery devices: %w", err)
	}

	var response struct {
		Data       []APINetworkDiscovery `json:"data"`
		Pagination struct {
			NextCursor string `json:"nextCursor"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse network discovery response: %w", err)
	}

	return &NetworkDiscoveryBatchResult{
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
		return nil, &httperr.APIError{Code: resp.StatusCode}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
