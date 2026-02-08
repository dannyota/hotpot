package targethttpproxy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeTargetHttpProxyWorkflowParams contains parameters for the target HTTP proxy workflow.
type GCPComputeTargetHttpProxyWorkflowParams struct {
	ProjectID string
}

// GCPComputeTargetHttpProxyWorkflowResult contains the result of the target HTTP proxy workflow.
type GCPComputeTargetHttpProxyWorkflowResult struct {
	ProjectID            string
	TargetHttpProxyCount int
	DurationMillis       int64
}

// GCPComputeTargetHttpProxyWorkflow ingests GCP Compute target HTTP proxies for a single project.
func GCPComputeTargetHttpProxyWorkflow(ctx workflow.Context, params GCPComputeTargetHttpProxyWorkflowParams) (*GCPComputeTargetHttpProxyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeTargetHttpProxyWorkflow", "projectID", params.ProjectID)

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

	var result IngestComputeTargetHttpProxiesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeTargetHttpProxiesActivity, IngestComputeTargetHttpProxiesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest target HTTP proxies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeTargetHttpProxyWorkflow",
		"projectID", params.ProjectID,
		"targetHttpProxyCount", result.TargetHttpProxyCount,
	)

	return &GCPComputeTargetHttpProxyWorkflowResult{
		ProjectID:            result.ProjectID,
		TargetHttpProxyCount: result.TargetHttpProxyCount,
		DurationMillis:       result.DurationMillis,
	}, nil
}
