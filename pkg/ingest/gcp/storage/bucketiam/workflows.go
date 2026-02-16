package bucketiam

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPStorageBucketIamWorkflowParams contains parameters for the bucket IAM workflow.
type GCPStorageBucketIamWorkflowParams struct {
	ProjectID string
}

// GCPStorageBucketIamWorkflowResult contains the result of the bucket IAM workflow.
type GCPStorageBucketIamWorkflowResult struct {
	ProjectID   string
	PolicyCount int
}

// GCPStorageBucketIamWorkflow ingests GCP Storage bucket IAM policies for a single project.
func GCPStorageBucketIamWorkflow(ctx workflow.Context, params GCPStorageBucketIamWorkflowParams) (*GCPStorageBucketIamWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPStorageBucketIamWorkflow", "projectID", params.ProjectID)

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

	var result IngestStorageBucketIamPoliciesResult
	err := workflow.ExecuteActivity(activityCtx, IngestStorageBucketIamPoliciesActivity, IngestStorageBucketIamPoliciesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest bucket IAM policies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPStorageBucketIamWorkflow",
		"projectID", params.ProjectID,
		"policyCount", result.PolicyCount,
	)

	return &GCPStorageBucketIamWorkflowResult{
		ProjectID:   result.ProjectID,
		PolicyCount: result.PolicyCount,
	}, nil
}
