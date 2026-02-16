package keyring

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPKMSKeyRingWorkflowParams contains parameters for the key ring workflow.
type GCPKMSKeyRingWorkflowParams struct {
	ProjectID string
}

// GCPKMSKeyRingWorkflowResult contains the result of the key ring workflow.
type GCPKMSKeyRingWorkflowResult struct {
	ProjectID      string
	KeyRingCount   int
	DurationMillis int64
}

// GCPKMSKeyRingWorkflow ingests GCP KMS key rings for a single project.
func GCPKMSKeyRingWorkflow(ctx workflow.Context, params GCPKMSKeyRingWorkflowParams) (*GCPKMSKeyRingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPKMSKeyRingWorkflow", "projectID", params.ProjectID)

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

	var result IngestKMSKeyRingsResult
	err := workflow.ExecuteActivity(activityCtx, IngestKMSKeyRingsActivity, IngestKMSKeyRingsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest key rings", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPKMSKeyRingWorkflow",
		"projectID", params.ProjectID,
		"keyRingCount", result.KeyRingCount,
	)

	return &GCPKMSKeyRingWorkflowResult{
		ProjectID:      result.ProjectID,
		KeyRingCount:   result.KeyRingCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
