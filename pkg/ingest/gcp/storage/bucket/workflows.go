package bucket

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPStorageBucketWorkflowParams contains parameters for the bucket workflow.
type GCPStorageBucketWorkflowParams struct {
	ProjectID string
}

// GCPStorageBucketWorkflowResult contains the result of the bucket workflow.
type GCPStorageBucketWorkflowResult struct {
	ProjectID      string
	BucketCount    int
	DurationMillis int64
}

// GCPStorageBucketWorkflow ingests GCP Storage buckets for a single project.
func GCPStorageBucketWorkflow(ctx workflow.Context, params GCPStorageBucketWorkflowParams) (*GCPStorageBucketWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPStorageBucketWorkflow", "projectID", params.ProjectID)

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

	var result IngestStorageBucketsResult
	err := workflow.ExecuteActivity(activityCtx, IngestStorageBucketsActivity, IngestStorageBucketsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest buckets", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPStorageBucketWorkflow",
		"projectID", params.ProjectID,
		"bucketCount", result.BucketCount,
	)

	return &GCPStorageBucketWorkflowResult{
		ProjectID:      result.ProjectID,
		BucketCount:    result.BucketCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
