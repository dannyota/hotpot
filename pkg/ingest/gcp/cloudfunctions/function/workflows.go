package function

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPCloudFunctionsFunctionWorkflowParams contains parameters for the Cloud Functions function workflow.
type GCPCloudFunctionsFunctionWorkflowParams struct {
	ProjectID string
}

// GCPCloudFunctionsFunctionWorkflowResult contains the result of the Cloud Functions function workflow.
type GCPCloudFunctionsFunctionWorkflowResult struct {
	ProjectID      string
	FunctionCount  int
	DurationMillis int64
}

// GCPCloudFunctionsFunctionWorkflow ingests GCP Cloud Functions for a single project.
func GCPCloudFunctionsFunctionWorkflow(ctx workflow.Context, params GCPCloudFunctionsFunctionWorkflowParams) (*GCPCloudFunctionsFunctionWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPCloudFunctionsFunctionWorkflow", "projectID", params.ProjectID)

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

	var result IngestCloudFunctionsFunctionsResult
	err := workflow.ExecuteActivity(activityCtx, IngestCloudFunctionsFunctionsActivity, IngestCloudFunctionsFunctionsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Cloud Functions", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPCloudFunctionsFunctionWorkflow",
		"projectID", params.ProjectID,
		"functionCount", result.FunctionCount,
	)

	return &GCPCloudFunctionsFunctionWorkflowResult{
		ProjectID:      result.ProjectID,
		FunctionCount:  result.FunctionCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
