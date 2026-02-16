package finding

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPSecurityCenterFindingWorkflowParams contains parameters for the SCC finding workflow.
type GCPSecurityCenterFindingWorkflowParams struct {
}

// GCPSecurityCenterFindingWorkflowResult contains the result of the SCC finding workflow.
type GCPSecurityCenterFindingWorkflowResult struct {
	FindingCount int
}

// GCPSecurityCenterFindingWorkflow ingests SCC findings.
func GCPSecurityCenterFindingWorkflow(ctx workflow.Context, params GCPSecurityCenterFindingWorkflowParams) (*GCPSecurityCenterFindingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSecurityCenterFindingWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestFindingsResult
	err := workflow.ExecuteActivity(activityCtx, IngestFindingsActivity, IngestFindingsParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest SCC findings", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPSecurityCenterFindingWorkflow",
		"findingCount", result.FindingCount,
	)

	return &GCPSecurityCenterFindingWorkflowResult{
		FindingCount: result.FindingCount,
	}, nil
}
