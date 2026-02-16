package instance

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPSpannerInstanceWorkflowParams contains parameters for the Spanner instance workflow.
type GCPSpannerInstanceWorkflowParams struct {
	ProjectID string
}

// GCPSpannerInstanceWorkflowResult contains the result of the Spanner instance workflow.
type GCPSpannerInstanceWorkflowResult struct {
	ProjectID      string
	InstanceCount  int
	InstanceNames  []string
	DurationMillis int64
}

// GCPSpannerInstanceWorkflow ingests GCP Spanner instances for a single project.
func GCPSpannerInstanceWorkflow(ctx workflow.Context, params GCPSpannerInstanceWorkflowParams) (*GCPSpannerInstanceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSpannerInstanceWorkflow", "projectID", params.ProjectID)

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

	var result IngestSpannerInstancesResult
	err := workflow.ExecuteActivity(activityCtx, IngestSpannerInstancesActivity, IngestSpannerInstancesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Spanner instances", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPSpannerInstanceWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
	)

	return &GCPSpannerInstanceWorkflowResult{
		ProjectID:      result.ProjectID,
		InstanceCount:  result.InstanceCount,
		InstanceNames:  result.InstanceNames,
		DurationMillis: result.DurationMillis,
	}, nil
}
