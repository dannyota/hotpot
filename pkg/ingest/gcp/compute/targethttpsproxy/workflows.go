package targethttpsproxy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeTargetHttpsProxyWorkflowParams contains parameters for the target HTTPS proxy workflow.
type GCPComputeTargetHttpsProxyWorkflowParams struct {
	ProjectID string
}

// GCPComputeTargetHttpsProxyWorkflowResult contains the result of the target HTTPS proxy workflow.
type GCPComputeTargetHttpsProxyWorkflowResult struct {
	ProjectID             string
	TargetHttpsProxyCount int
	DurationMillis        int64
}

// GCPComputeTargetHttpsProxyWorkflow ingests GCP Compute target HTTPS proxies for a single project.
func GCPComputeTargetHttpsProxyWorkflow(ctx workflow.Context, params GCPComputeTargetHttpsProxyWorkflowParams) (*GCPComputeTargetHttpsProxyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeTargetHttpsProxyWorkflow", "projectID", params.ProjectID)

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

	var result IngestComputeTargetHttpsProxiesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeTargetHttpsProxiesActivity, IngestComputeTargetHttpsProxiesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest target HTTPS proxies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeTargetHttpsProxyWorkflow",
		"projectID", params.ProjectID,
		"targetHttpsProxyCount", result.TargetHttpsProxyCount,
	)

	return &GCPComputeTargetHttpsProxyWorkflowResult{
		ProjectID:             result.ProjectID,
		TargetHttpsProxyCount: result.TargetHttpsProxyCount,
		DurationMillis:        result.DurationMillis,
	}, nil
}
