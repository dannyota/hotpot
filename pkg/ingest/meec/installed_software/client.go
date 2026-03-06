package installed_software

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"danny.vn/hotpot/pkg/base/httperr"
)

// Client wraps the MEEC Inventory Installed Software API.
type Client struct {
	baseURL    string
	apiToken   string
	apiVersion string
	httpClient *http.Client
}

// NewClient creates a new MEEC installed software client.
func NewClient(baseURL, apiToken, apiVersion string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		apiVersion: apiVersion,
		httpClient: httpClient,
	}
}

// APIInstalledSoftware represents installed software from the MEEC API response.
// Fields with value "--" represent empty/null.
type APIInstalledSoftware struct {
	SoftwareID       int         `json:"software_id"`
	SoftwareName     string      `json:"software_name"`
	SoftwareVersion  string      `json:"software_version"`
	DisplayName      any `json:"display_name"`
	ManufacturerName any `json:"manufacturer_name"`
	InstalledDate    int64       `json:"installed_date"`
	Architecture     any `json:"architecture"`
	Location         any `json:"location"`
	SwType           int         `json:"sw_type"`
	SwCategoryName   any `json:"sw_category_name"`
	DetectedTime     int64       `json:"detected_time"`
	ManufacturerID   int         `json:"manufacturer_id"`
	ManagedSwID      int         `json:"managed_sw_id"`
	InstalledFormat  any `json:"installed_format"`
	IsUsageProhibit  int         `json:"is_usage_prohibited"`
	Comments         any `json:"comments"`
	CompliantStatus  any `json:"compliant_status"`
	TotalCopies      any `json:"total_copies"`
	RemainingCopies  any `json:"remaining_copies"`
	SwFamily         int         `json:"sw_family"`
}

// GetInstalledSoftware retrieves all installed software for a single computer.
func (c *Client) GetInstalledSoftware(resourceID string) ([]APIInstalledSoftware, error) {
	endpoint := fmt.Sprintf("/api/%s/inventory/installedsoftware", c.apiVersion)
	requestURL := fmt.Sprintf("%s%s?resid=%s", c.baseURL, endpoint, resourceID)

	start := time.Now()

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	slog.Debug("meec installed software api request", "endpoint", endpoint, "resourceID", resourceID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("meec installed software api request failed", "endpoint", endpoint, "resourceID", resourceID, "error", err, "durationMs", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	slog.Debug("meec installed software api response", "endpoint", endpoint, "resourceID", resourceID, "status", resp.StatusCode, "responseBytes", len(body), "durationMs", time.Since(start).Milliseconds())

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, &httperr.APIError{Code: resp.StatusCode}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		MessageResponse struct {
			InstalledSoftware []APIInstalledSoftware `json:"installedsoftware"`
		} `json:"message_response"`
		Status           string `json:"status"`
		ErrorCode        string `json:"error_code"`
		ErrorDescription string `json:"error_description"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse installed software response: %w", err)
	}

	// MEEC returns errors as HTTP 200 with "status": "error".
	if response.Status == "error" {
		slog.Error("meec api error", "endpoint", endpoint, "resourceID", resourceID, "errorCode", response.ErrorCode, "errorDescription", response.ErrorDescription)
		if response.ErrorCode == "10002" {
			return nil, &httperr.APIError{Code: http.StatusUnauthorized}
		}
		return nil, fmt.Errorf("MEEC API error %s: %s", response.ErrorCode, response.ErrorDescription)
	}

	return response.MessageResponse.InstalledSoftware, nil
}
