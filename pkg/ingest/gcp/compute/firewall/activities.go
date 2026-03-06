package firewall

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"

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
	client         *Client
	entClient      *entcompute.Client
	temporalClient client.Client
}

// NewActivities creates a new Activities instance.
func NewActivities(client *Client, entClient *entcompute.Client, temporalClient client.Client) *Activities {
	return &Activities{
		client:         client,
		entClient:      entClient,
		temporalClient: temporalClient,
	}
}

// FetchAndSaveFirewallsPageParams contains parameters for the page activity.
type FetchAndSaveFirewallsPageParams struct {
	ProjectID   string
	PageToken   string
	PageSize    int
	CollectedAt time.Time
	WorkflowID  string
	RunID       string
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

	firewalls, nextToken, err := a.client.ListFirewallsPage(ctx, params.ProjectID, params.PageSize, params.PageToken)
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
