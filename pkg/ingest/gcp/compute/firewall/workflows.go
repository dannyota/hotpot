package firewall

import (
	"time"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const pageSize = 500

// GCPComputeFirewallWorkflowParams contains parameters for the firewall workflow.
type GCPComputeFirewallWorkflowParams struct {
	ProjectID string
}

// GCPComputeFirewallWorkflowResult contains the result of the firewall workflow.
type GCPComputeFirewallWorkflowResult struct {
	ProjectID     string
	FirewallCount int
}

// GCPComputeFirewallWorkflow ingests GCP Compute firewalls for a single project.
// It pipelines page fetches: as each activity finishes its GCP fetch, it signals
// the workflow with the next page token, and the workflow immediately starts the
// next activity. All activities run concurrently (rate-limited by the GCP limiter).
func GCPComputeFirewallWorkflow(ctx workflow.Context, params GCPComputeFirewallWorkflowParams) (*GCPComputeFirewallWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeFirewallWorkflow", "projectID", params.ProjectID)

	collectedAt := workflow.Now(ctx)
	wfInfo := workflow.GetInfo(ctx).WorkflowExecution

	fetchOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	fetchCtx := workflow.WithActivityOptions(ctx, fetchOpts)

	signalCh := workflow.GetSignalChannel(ctx, firewallNextPageSignal)

	// Start first page activity.
	latestFuture := workflow.ExecuteActivity(fetchCtx, FetchAndSaveFirewallsPageActivity, FetchAndSaveFirewallsPageParams{
		ProjectID:   params.ProjectID,
		PageToken:   "",
		PageSize:    pageSize,
		CollectedAt: collectedAt,
		WorkflowID:  wfInfo.ID,
		RunID:       wfInfo.RunID,
	})
	allFutures := []workflow.Future{latestFuture}

	// Dispatch loop: fire new activities as signals arrive.
	for {
		var signal NextPageTokenSignal
		var gotSignal bool

		selector := workflow.NewSelector(ctx)
		selector.AddReceive(signalCh, func(ch workflow.ReceiveChannel, _ bool) {
			ch.Receive(ctx, &signal)
			gotSignal = true
		})
		selector.AddFuture(latestFuture, func(f workflow.Future) {
			// Latest activity completed before signaling.
		})
		selector.Select(ctx)

		if !gotSignal {
			// Latest activity finished without signaling — failed during fetch
			// or signal delivery failed. Check the result.
			var result FetchAndSaveFirewallsPageResult
			if err := latestFuture.Get(ctx, &result); err != nil {
				return nil, temporalerr.PropagateNonRetryable(err)
			}
			// Signal delivery failed — fall back to using token from result.
			if result.NextPageToken == "" {
				break
			}
			latestFuture = workflow.ExecuteActivity(fetchCtx, FetchAndSaveFirewallsPageActivity, FetchAndSaveFirewallsPageParams{
				ProjectID:   params.ProjectID,
				PageToken:   result.NextPageToken,
				PageSize:    pageSize,
				CollectedAt: collectedAt,
				WorkflowID:  wfInfo.ID,
				RunID:       wfInfo.RunID,
			})
			allFutures = append(allFutures, latestFuture)
			continue
		}

		if signal.NextPageToken == "" {
			break // Last page — no more activities to start.
		}

		latestFuture = workflow.ExecuteActivity(fetchCtx, FetchAndSaveFirewallsPageActivity, FetchAndSaveFirewallsPageParams{
			ProjectID:   params.ProjectID,
			PageToken:   signal.NextPageToken,
			PageSize:    pageSize,
			CollectedAt: collectedAt,
			WorkflowID:  wfInfo.ID,
			RunID:       wfInfo.RunID,
		})
		allFutures = append(allFutures, latestFuture)
	}

	// Wait for all page activities to complete.
	totalCount := 0
	for _, f := range allFutures {
		var result FetchAndSaveFirewallsPageResult
		if err := f.Get(ctx, &result); err != nil {
			return nil, temporalerr.PropagateNonRetryable(err)
		}
		totalCount += result.Count
	}

	// Delete stale firewalls.
	deleteOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	deleteCtx := workflow.WithActivityOptions(ctx, deleteOpts)

	err := workflow.ExecuteActivity(deleteCtx, DeleteStaleFirewallsActivity, DeleteStaleFirewallsParams{
		ProjectID:   params.ProjectID,
		CollectedAt: collectedAt,
	}).Get(ctx, nil)
	if err != nil {
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed GCPComputeFirewallWorkflow",
		"projectID", params.ProjectID,
		"firewallCount", totalCount,
	)

	return &GCPComputeFirewallWorkflowResult{
		ProjectID:     params.ProjectID,
		FirewallCount: totalCount,
	}, nil
}
