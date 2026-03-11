package interconnect

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeNetworkInterconnectWorkflowParams contains parameters for the interconnect workflow.
type GreenNodeNetworkInterconnectWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeNetworkInterconnectWorkflowResult contains the result of the interconnect workflow.
type GreenNodeNetworkInterconnectWorkflowResult struct {
	InterconnectCount int
	DurationMillis    int64
}

// GreenNodeNetworkInterconnectWorkflow ingests GreenNode interconnects.
func GreenNodeNetworkInterconnectWorkflow(ctx workflow.Context, params GreenNodeNetworkInterconnectWorkflowParams) (*GreenNodeNetworkInterconnectWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeNetworkInterconnectWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestNetworkInterconnectsResult
	err := workflow.ExecuteActivity(activityCtx, IngestNetworkInterconnectsActivity, IngestNetworkInterconnectsParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest interconnects", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeNetworkInterconnectWorkflow", "interconnectCount", result.InterconnectCount)

	return &GreenNodeNetworkInterconnectWorkflowResult{
		InterconnectCount: result.InterconnectCount,
		DurationMillis:    result.DurationMillis,
	}, nil
}
