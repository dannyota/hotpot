package storage

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/storage/bucket"
)

// GCPStorageWorkflowParams contains parameters for the storage workflow.
type GCPStorageWorkflowParams struct {
	ProjectID string
}

// GCPStorageWorkflowResult contains the result of the storage workflow.
type GCPStorageWorkflowResult struct {
	ProjectID   string
	BucketCount int
}

// GCPStorageWorkflow ingests all GCP Storage resources for a single project.
func GCPStorageWorkflow(ctx workflow.Context, params GCPStorageWorkflowParams) (*GCPStorageWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPStorageWorkflow", "projectID", params.ProjectID)

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPStorageWorkflowResult{
		ProjectID: params.ProjectID,
	}

	var bucketResult bucket.GCPStorageBucketWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, bucket.GCPStorageBucketWorkflow,
		bucket.GCPStorageBucketWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &bucketResult)
	if err != nil {
		logger.Error("Failed to ingest buckets", "error", err)
		return nil, err
	}
	result.BucketCount = bucketResult.BucketCount

	logger.Info("Completed GCPStorageWorkflow",
		"projectID", params.ProjectID,
		"bucketCount", result.BucketCount,
	)

	return result, nil
}
