package logbucket

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPLoggingBucketWorkflowParams contains parameters for the bucket workflow.
type GCPLoggingBucketWorkflowParams struct {
	ProjectID string
}

// GCPLoggingBucketWorkflowResult contains the result of the bucket workflow.
type GCPLoggingBucketWorkflowResult struct {
	ProjectID      string
	BucketCount    int
	DurationMillis int64
}

// GCPLoggingBucketWorkflow ingests GCP Cloud Logging buckets for a single project.
func GCPLoggingBucketWorkflow(ctx workflow.Context, params GCPLoggingBucketWorkflowParams) (*GCPLoggingBucketWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPLoggingBucketWorkflow", "projectID", params.ProjectID)

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

	var result IngestLoggingBucketsResult
	err := workflow.ExecuteActivity(activityCtx, IngestLoggingBucketsActivity, IngestLoggingBucketsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest buckets", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPLoggingBucketWorkflow",
		"projectID", params.ProjectID,
		"bucketCount", result.BucketCount,
	)

	return &GCPLoggingBucketWorkflowResult{
		ProjectID:      result.ProjectID,
		BucketCount:    result.BucketCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
