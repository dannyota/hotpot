package folder

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPResourceManagerFolderWorkflowParams contains parameters for the folder workflow.
type GCPResourceManagerFolderWorkflowParams struct {
	// Empty - discovers all accessible folders
}

// GCPResourceManagerFolderWorkflowResult contains the result of the folder workflow.
type GCPResourceManagerFolderWorkflowResult struct {
	FolderCount    int
	FolderIDs      []string
	DurationMillis int64
}

// GCPResourceManagerFolderWorkflow discovers all GCP folders accessible by the service account.
func GCPResourceManagerFolderWorkflow(ctx workflow.Context, params GCPResourceManagerFolderWorkflowParams) (*GCPResourceManagerFolderWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPResourceManagerFolderWorkflow")

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
	var result IngestFoldersResult
	err := workflow.ExecuteActivity(activityCtx, IngestFoldersActivity, IngestFoldersParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to discover folders", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPResourceManagerFolderWorkflow",
		"folderCount", result.FolderCount,
	)

	return &GCPResourceManagerFolderWorkflowResult{
		FolderCount:    result.FolderCount,
		FolderIDs:      result.FolderIDs,
		DurationMillis: result.DurationMillis,
	}, nil
}
