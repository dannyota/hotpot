package neg

import (
	"time"

	"danny.vn/hotpot/pkg/base/temporalerr"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type GCPComputeNegWorkflowParams struct {
	ProjectID      string
	QuotaProjectID string
}

type GCPComputeNegWorkflowResult struct {
	ProjectID string
	NegCount  int
}

func GCPComputeNegWorkflow(ctx workflow.Context, params GCPComputeNegWorkflowParams) (*GCPComputeNegWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeNegWorkflow", "projectID", params.ProjectID)

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

	var result IngestComputeNegsResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeNegsActivity,
		IngestComputeNegsParams{ProjectID: params.ProjectID, QuotaProjectID: params.QuotaProjectID}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest NEGs", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed GCPComputeNegWorkflow",
		"projectID", params.ProjectID,
		"negCount", result.NegCount,
	)

	return &GCPComputeNegWorkflowResult{
		ProjectID: result.ProjectID,
		NegCount:  result.NegCount,
	}, nil
}
