package source

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPSecurityCenterSourceWorkflowParams contains parameters for the SCC source workflow.
type GCPSecurityCenterSourceWorkflowParams struct {
}

// GCPSecurityCenterSourceWorkflowResult contains the result of the SCC source workflow.
type GCPSecurityCenterSourceWorkflowResult struct {
	SourceCount int
}

// GCPSecurityCenterSourceWorkflow ingests SCC sources.
func GCPSecurityCenterSourceWorkflow(ctx workflow.Context, params GCPSecurityCenterSourceWorkflowParams) (*GCPSecurityCenterSourceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSecurityCenterSourceWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestSourcesResult
	err := workflow.ExecuteActivity(activityCtx, IngestSourcesActivity, IngestSourcesParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest SCC sources", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPSecurityCenterSourceWorkflow",
		"sourceCount", result.SourceCount,
	)

	return &GCPSecurityCenterSourceWorkflowResult{
		SourceCount: result.SourceCount,
	}, nil
}
