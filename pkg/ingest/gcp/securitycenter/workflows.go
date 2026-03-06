package securitycenter

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
	"danny.vn/hotpot/pkg/ingest/gcp/securitycenter/finding"
	"danny.vn/hotpot/pkg/ingest/gcp/securitycenter/notificationconfig"
	"danny.vn/hotpot/pkg/ingest/gcp/securitycenter/source"
)

// GCPSecurityCenterWorkflowParams contains parameters for the SCC workflow.
type GCPSecurityCenterWorkflowParams struct {
}

// GCPSecurityCenterWorkflowResult contains the result of the SCC workflow.
type GCPSecurityCenterWorkflowResult struct {
	SourceCount             int
	FindingCount            int
	NotificationConfigCount int
}

// GCPSecurityCenterWorkflow ingests all Security Command Center resources.
// Sources run first (findings reference sources), then findings and notification configs run in parallel.
func GCPSecurityCenterWorkflow(ctx workflow.Context, params GCPSecurityCenterWorkflowParams) (*GCPSecurityCenterWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSecurityCenterWorkflow")

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 60 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPSecurityCenterWorkflowResult{}

	// Phase 1: Ingest sources first (findings reference sources)
	var sourceResult source.GCPSecurityCenterSourceWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, source.GCPSecurityCenterSourceWorkflow,
		source.GCPSecurityCenterSourceWorkflowParams{}).Get(ctx, &sourceResult)
	if err != nil {
		logger.Error("Failed to ingest SCC sources", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}
	result.SourceCount = sourceResult.SourceCount

	// Phase 2: Ingest findings and notification configs in parallel (both independent after sources)
	findingFuture := workflow.ExecuteChildWorkflow(childCtx, finding.GCPSecurityCenterFindingWorkflow,
		finding.GCPSecurityCenterFindingWorkflowParams{})

	notificationConfigFuture := workflow.ExecuteChildWorkflow(childCtx, notificationconfig.GCPSecurityCenterNotificationConfigWorkflow,
		notificationconfig.GCPSecurityCenterNotificationConfigWorkflowParams{})

	// Collect finding results
	var findingResult finding.GCPSecurityCenterFindingWorkflowResult
	if err := findingFuture.Get(ctx, &findingResult); err != nil {
		logger.Error("Failed to ingest SCC findings", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}
	result.FindingCount = findingResult.FindingCount

	// Collect notification config results
	var notificationConfigResult notificationconfig.GCPSecurityCenterNotificationConfigWorkflowResult
	if err := notificationConfigFuture.Get(ctx, &notificationConfigResult); err != nil {
		logger.Error("Failed to ingest SCC notification configs", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}
	result.NotificationConfigCount = notificationConfigResult.NotificationConfigCount

	logger.Info("Completed GCPSecurityCenterWorkflow",
		"sourceCount", result.SourceCount,
		"findingCount", result.FindingCount,
		"notificationConfigCount", result.NotificationConfigCount,
	)

	return result, nil
}
