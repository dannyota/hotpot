package software

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"danny.vn/hotpot/pkg/base/httperr"
)

// Client wraps the MEEC Inventory Software API.
type Client struct {
	baseURL    string
	apiToken   string
	apiVersion string
	httpClient *http.Client
}

// NewClient creates a new MEEC software client.
func NewClient(baseURL, apiToken, apiVersion string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		apiVersion: apiVersion,
		httpClient: httpClient,
	}
}

// APISoftware represents a software entry from the MEEC inventory API.
type APISoftware struct {
	SoftwareID            int    `json:"software_id"`
	SoftwareName          string `json:"software_name"`
	SoftwareVersion       string `json:"software_version"`
	DisplayName           string `json:"display_name"`
	ManufacturerID        int    `json:"manufacturer_id"`
	ManufacturerName      string `json:"manufacturer_name"`
	SwCategoryName        string `json:"sw_category_name"`
	SwType                int    `json:"sw_type"`
	SwFamily              int    `json:"sw_family"`
	InstalledFormat       string `json:"installed_format"`
	IsUsageProhibited     int    `json:"is_usage_prohibited"`
	ManagedInstallations  int    `json:"managed_installations"`
	NetworkInstallations  int    `json:"network_installations"`
	ManagedSwID           int    `json:"managed_sw_id"`
	DetectedTime          int64  `json:"detected_time"`
	CompliantStatus       string `json:"compliant_status"`
	TotalCopies           string `json:"total_copies"`
	RemainingCopies       string `json:"remaining_copies"`
	Comments              string `json:"comments"`
}

// SoftwareBatchResult contains a batch of software entries and pagination info.
type SoftwareBatchResult struct {
	Software []APISoftware
	Total    int
	Page     int
	HasMore  bool
}

const pageLimit = 1000

// GetSoftwareBatch retrieves a page of software entries.
func (c *Client) GetSoftwareBatch(page int) (*SoftwareBatchResult, error) {
	params := url.Values{}
	params.Set("pagelimit", fmt.Sprintf("%d", pageLimit))
	params.Set("page", fmt.Sprintf("%d", page))

	body, err := c.doRequest("GET", fmt.Sprintf("/api/%s/inventory/software", c.apiVersion), params)
	if err != nil {
		return nil, fmt.Errorf("get software batch: %w", err)
	}

	var response struct {
		MessageResponse struct {
			Software []APISoftware `json:"software"`
			Total    int           `json:"total"`
			Limit    int           `json:"limit"`
			Page     int           `json:"page"`
		} `json:"message_response"`
		Status string `json:"status"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse software response: %w", err)
	}

	fetched := response.MessageResponse.Page * response.MessageResponse.Limit
	hasMore := fetched < response.MessageResponse.Total

	return &SoftwareBatchResult{
		Software: response.MessageResponse.Software,
		Total:    response.MessageResponse.Total,
		Page:     response.MessageResponse.Page,
		HasMore:  hasMore,
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

	req.Header.Set("Authorization", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	slog.Debug("meec api request", "method", method, "endpoint", endpoint)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("meec api request failed", "method", method, "endpoint", endpoint, "error", err, "durationMs", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	slog.Info("meec api response", "method", method, "endpoint", endpoint, "status", resp.StatusCode, "responseBytes", len(body), "durationMs", time.Since(start).Milliseconds())

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, &httperr.APIError{Code: resp.StatusCode}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// MEEC returns errors as HTTP 200 with "status": "error".
	var envelope struct {
		Status           string `json:"status"`
		ErrorCode        string `json:"error_code"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Status == "error" {
		slog.Error("meec api error", "method", method, "endpoint", endpoint, "errorCode", envelope.ErrorCode, "errorDescription", envelope.ErrorDescription)
		if envelope.ErrorCode == "10002" {
			return nil, &httperr.APIError{Code: http.StatusUnauthorized}
		}
		return nil, fmt.Errorf("MEEC API error %s: %s", envelope.ErrorCode, envelope.ErrorDescription)
	}

	return body, nil
}
