package firewall

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"google.golang.org/api/option"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/gcpauth"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
)

const firewallNextPageSignal = "firewall-next-page"

// NextPageTokenSignal is the signal payload sent from activity to workflow.
type NextPageTokenSignal struct {
	NextPageToken string
}

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService  *config.Service
	entClient      *entcompute.Client
	limiter        ratelimit.Limiter
	temporalClient client.Client
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entcompute.Client, limiter ratelimit.Limiter, temporalClient client.Client) *Activities {
	return &Activities{
		configService:  configService,
		entClient:      entClient,
		limiter:        limiter,
		temporalClient: temporalClient,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context, quotaProjectID string) (*Client, error) {
	httpClient, err := gcpauth.NewHTTPClient(ctx, a.configService.GCPCredentialsJSON(), a.limiter)
	if err != nil {
		return nil, err
	}
	opts := []option.ClientOption{option.WithHTTPClient(httpClient)}
	if quotaProjectID != "" {
		opts = append(opts, option.WithQuotaProject(quotaProjectID))
	}
	return NewClient(ctx, opts...)
}

// FetchAndSaveFirewallsPageParams contains parameters for the page activity.
type FetchAndSaveFirewallsPageParams struct {
	ProjectID      string
	QuotaProjectID string
	PageToken      string
	PageSize       int
	CollectedAt    time.Time
	WorkflowID     string
	RunID          string
}

// FetchAndSaveFirewallsPageResult contains the result of the page activity.
type FetchAndSaveFirewallsPageResult struct {
	Count         int
	NextPageToken string
}

// FetchAndSaveFirewallsPageActivity is the activity function reference for workflow registration.
var FetchAndSaveFirewallsPageActivity = (*Activities).FetchAndSaveFirewallsPage

// FetchAndSaveFirewallsPage fetches one page of firewalls from GCP and saves to DB.
// After the GCP fetch, it signals the workflow with the next page token so the
// workflow can start the next fetch while this activity saves to DB.
func (a *Activities) FetchAndSaveFirewallsPage(ctx context.Context, params FetchAndSaveFirewallsPageParams) (*FetchAndSaveFirewallsPageResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Fetching firewall page",
		"projectID", params.ProjectID,
		"pageSize", params.PageSize,
		"hasPageToken", params.PageToken != "",
	)

	// Create client for this activity
	gcpClient, err := a.createClient(ctx, params.QuotaProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer gcpClient.Close()

	firewalls, nextToken, err := gcpClient.ListFirewallsPage(ctx, params.ProjectID, params.PageSize, params.PageToken)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("list firewalls page: %w", err))
	}

	// Signal workflow with next page token BEFORE saving to DB.
	// This lets the workflow start the next GCP fetch while we save.
	if err := a.temporalClient.SignalWorkflow(ctx, params.WorkflowID, params.RunID,
		firewallNextPageSignal, NextPageTokenSignal{NextPageToken: nextToken}); err != nil {
		logger.Warn("Failed to signal next page token", "error", err)
		// Non-fatal: workflow will fall back to reading token from result.
	}

	service := NewService(nil, a.entClient)
	count, err := service.IngestPage(ctx, firewalls, params.ProjectID, params.CollectedAt)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest firewalls page: %w", err))
	}

	logger.Info("Saved firewall page",
		"projectID", params.ProjectID,
		"count", count,
		"hasNextPage", nextToken != "",
	)

	return &FetchAndSaveFirewallsPageResult{
		Count:         count,
		NextPageToken: nextToken,
	}, nil
}

// DeleteStaleFirewallsParams contains parameters for the delete stale activity.
type DeleteStaleFirewallsParams struct {
	ProjectID   string
	CollectedAt time.Time
}

// DeleteStaleFirewallsActivity is the activity function reference for workflow registration.
var DeleteStaleFirewallsActivity = (*Activities).DeleteStaleFirewalls

// DeleteStaleFirewalls removes firewalls not seen in this collection run.
func (a *Activities) DeleteStaleFirewalls(ctx context.Context, params DeleteStaleFirewallsParams) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Deleting stale firewalls", "projectID", params.ProjectID)

	service := NewService(nil, a.entClient)
	if err := service.DeleteStaleFirewalls(ctx, params.ProjectID, params.CollectedAt); err != nil {
		return temporalerr.MaybeNonRetryable(fmt.Errorf("delete stale firewalls: %w", err))
	}

	return nil
}
