package sshkey

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeComputeSSHKeyWorkflowParams contains parameters for the SSH key workflow.
type GreenNodeComputeSSHKeyWorkflowParams struct {
	ProjectID string
}

// GreenNodeComputeSSHKeyWorkflowResult contains the result of the SSH key workflow.
type GreenNodeComputeSSHKeyWorkflowResult struct {
	KeyCount       int
	DurationMillis int64
}

// GreenNodeComputeSSHKeyWorkflow ingests GreenNode SSH keys.
func GreenNodeComputeSSHKeyWorkflow(ctx workflow.Context, params GreenNodeComputeSSHKeyWorkflowParams) (*GreenNodeComputeSSHKeyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeComputeSSHKeyWorkflow", "projectID", params.ProjectID)

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

	var result IngestComputeSSHKeysResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeSSHKeysActivity, IngestComputeSSHKeysParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest SSH keys", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeComputeSSHKeyWorkflow", "keyCount", result.KeyCount)

	return &GreenNodeComputeSSHKeyWorkflowResult{
		KeyCount:       result.KeyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
