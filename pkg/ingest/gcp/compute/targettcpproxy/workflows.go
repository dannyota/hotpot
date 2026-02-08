package targettcpproxy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeTargetTcpProxyWorkflowParams contains parameters for the target TCP proxy workflow.
type GCPComputeTargetTcpProxyWorkflowParams struct {
	ProjectID string
}

// GCPComputeTargetTcpProxyWorkflowResult contains the result of the target TCP proxy workflow.
type GCPComputeTargetTcpProxyWorkflowResult struct {
	ProjectID           string
	TargetTcpProxyCount int
	DurationMillis      int64
}

// GCPComputeTargetTcpProxyWorkflow ingests GCP Compute target TCP proxies for a single project.
func GCPComputeTargetTcpProxyWorkflow(ctx workflow.Context, params GCPComputeTargetTcpProxyWorkflowParams) (*GCPComputeTargetTcpProxyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeTargetTcpProxyWorkflow", "projectID", params.ProjectID)

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

	var result IngestComputeTargetTcpProxiesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeTargetTcpProxiesActivity, IngestComputeTargetTcpProxiesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest target TCP proxies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeTargetTcpProxyWorkflow",
		"projectID", params.ProjectID,
		"targetTcpProxyCount", result.TargetTcpProxyCount,
	)

	return &GCPComputeTargetTcpProxyWorkflowResult{
		ProjectID:           result.ProjectID,
		TargetTcpProxyCount: result.TargetTcpProxyCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
