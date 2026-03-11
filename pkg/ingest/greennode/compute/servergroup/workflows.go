package servergroup

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeComputeServerGroupWorkflowParams contains parameters for the server group workflow.
type GreenNodeComputeServerGroupWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeComputeServerGroupWorkflowResult contains the result of the server group workflow.
type GreenNodeComputeServerGroupWorkflowResult struct {
	GroupCount     int
	DurationMillis int64
}

// GreenNodeComputeServerGroupWorkflow ingests GreenNode server groups.
func GreenNodeComputeServerGroupWorkflow(ctx workflow.Context, params GreenNodeComputeServerGroupWorkflowParams) (*GreenNodeComputeServerGroupWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeComputeServerGroupWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestComputeServerGroupsResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeServerGroupsActivity, IngestComputeServerGroupsParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest server groups", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeComputeServerGroupWorkflow", "groupCount", result.GroupCount)

	return &GreenNodeComputeServerGroupWorkflowResult{
		GroupCount:     result.GroupCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
