package targetsslproxy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeTargetSslProxyWorkflowParams contains parameters for the target SSL proxy workflow.
type GCPComputeTargetSslProxyWorkflowParams struct {
	ProjectID string
}

// GCPComputeTargetSslProxyWorkflowResult contains the result of the target SSL proxy workflow.
type GCPComputeTargetSslProxyWorkflowResult struct {
	ProjectID           string
	TargetSslProxyCount int
	DurationMillis      int64
}

// GCPComputeTargetSslProxyWorkflow ingests GCP Compute target SSL proxies for a single project.
func GCPComputeTargetSslProxyWorkflow(ctx workflow.Context, params GCPComputeTargetSslProxyWorkflowParams) (*GCPComputeTargetSslProxyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeTargetSslProxyWorkflow", "projectID", params.ProjectID)

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

	var result IngestComputeTargetSslProxiesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeTargetSslProxiesActivity, IngestComputeTargetSslProxiesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest target SSL proxies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeTargetSslProxyWorkflow",
		"projectID", params.ProjectID,
		"targetSslProxyCount", result.TargetSslProxyCount,
	)

	return &GCPComputeTargetSslProxyWorkflowResult{
		ProjectID:           result.ProjectID,
		TargetSslProxyCount: result.TargetSslProxyCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
