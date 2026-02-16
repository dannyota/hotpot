package projectmetadata

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeProjectMetadataWorkflowParams contains parameters for the project metadata workflow.
type GCPComputeProjectMetadataWorkflowParams struct {
	ProjectID string
}

// GCPComputeProjectMetadataWorkflowResult contains the result of the project metadata workflow.
type GCPComputeProjectMetadataWorkflowResult struct {
	ProjectID      string
	MetadataCount  int
	DurationMillis int64
}

// GCPComputeProjectMetadataWorkflow ingests GCP Compute project metadata for a single project.
func GCPComputeProjectMetadataWorkflow(ctx workflow.Context, params GCPComputeProjectMetadataWorkflowParams) (*GCPComputeProjectMetadataWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeProjectMetadataWorkflow", "projectID", params.ProjectID)

	// Activity options
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

	// Execute ingest activity
	var result IngestComputeProjectMetadataResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeProjectMetadataActivity, IngestComputeProjectMetadataParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest project metadata", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeProjectMetadataWorkflow",
		"projectID", params.ProjectID,
		"metadataCount", result.MetadataCount,
	)

	return &GCPComputeProjectMetadataWorkflowResult{
		ProjectID:      result.ProjectID,
		MetadataCount:  result.MetadataCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
