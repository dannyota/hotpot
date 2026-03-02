package computer

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

// Client wraps the MEEC SOM Computers API.
type Client struct {
	baseURL    string
	apiToken   string
	apiVersion string
	httpClient *http.Client
}

// NewClient creates a new MEEC computers client.
func NewClient(baseURL, apiToken, apiVersion string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiToken:   apiToken,
		apiVersion: apiVersion,
		httpClient: httpClient,
	}
}

// APIComputer represents a computer from the MEEC API response.
// Many fields can be either their typed value or the string "--" meaning empty/unknown.
type APIComputer struct {
	ResourceID           int         `json:"resource_id"`
	ResourceName         string      `json:"resource_name"`
	FQDNName             any `json:"fqdn_name"`
	DomainNetbiosName    any `json:"domain_netbios_name"`
	IPAddress            any `json:"ip_address"`
	MACAddress           any `json:"mac_address"`
	OsName               any `json:"os_name"`
	OsPlatform           any `json:"os_platform"`
	OsPlatformName       any `json:"os_platform_name"`
	OsVersion            any `json:"os_version"`
	ServicePack          any `json:"service_pack"`
	AgentVersion         any `json:"agent_version"`
	ComputerLiveStatus   any `json:"computer_live_status"`
	InstallationStatus   any `json:"installation_status"`
	ManagedStatus        any `json:"managed_status"`
	BranchOfficeName     any `json:"branch_office_name"`
	Owner                any `json:"owner"`
	OwnerEmailID         any `json:"owner_email_id"`
	Description          any `json:"description"`
	Location             any `json:"location"`
	LastSyncTime         any `json:"last_sync_time"`
	AgentLastContactTime any `json:"agent_last_contact_time"`
	AgentInstalledOn     any `json:"agent_installed_on"`
	CustomerName         any `json:"customer_name"`
	CustomerID           any `json:"customer_id"`
}

// ComputerBatchResult contains a batch of computers and pagination info.
type ComputerBatchResult struct {
	Computers []APIComputer
	Total     int
	Page      int
	Limit     int
}

const defaultPageLimit = 1000

// GetAllComputers retrieves all computers using page-based pagination.
func (c *Client) GetAllComputers() ([]APIComputer, error) {
	var allComputers []APIComputer
	page := 1

	for {
		batch, err := c.getComputersBatch(page)
		if err != nil {
			return nil, fmt.Errorf("get computers page %d: %w", page, err)
		}

		allComputers = append(allComputers, batch.Computers...)

		slog.Info("meec computers batch fetched",
			"page", page,
			"batchItems", len(batch.Computers),
			"totalFetched", len(allComputers),
			"totalExpected", batch.Total,
		)

		if len(allComputers) >= batch.Total {
			break
		}
		page++
	}

	return allComputers, nil
}

func (c *Client) getComputersBatch(page int) (*ComputerBatchResult, error) {
	params := url.Values{}
	params.Set("pagelimit", fmt.Sprintf("%d", defaultPageLimit))
	params.Set("page", fmt.Sprintf("%d", page))

	endpoint := fmt.Sprintf("/api/%s/som/computers", c.apiVersion)
	body, err := c.doRequest("GET", endpoint, params)
	if err != nil {
		return nil, err
	}

	var response struct {
		MessageResponse struct {
			Computers []APIComputer `json:"computers"`
			Total     int           `json:"total"`
			Limit     int           `json:"limit"`
			Page      int           `json:"page"`
		} `json:"message_response"`
		Status string `json:"status"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse computers response: %w", err)
	}

	return &ComputerBatchResult{
		Computers: response.MessageResponse.Computers,
		Total:     response.MessageResponse.Total,
		Page:      response.MessageResponse.Page,
		Limit:     response.MessageResponse.Limit,
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
