package serviceperimeter

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPAccessContextManagerServicePerimeterWorkflowParams contains parameters for the service perimeter workflow.
type GCPAccessContextManagerServicePerimeterWorkflowParams struct {
}

// GCPAccessContextManagerServicePerimeterWorkflowResult contains the result of the service perimeter workflow.
type GCPAccessContextManagerServicePerimeterWorkflowResult struct {
	PerimeterCount int
}

// GCPAccessContextManagerServicePerimeterWorkflow ingests service perimeters.
func GCPAccessContextManagerServicePerimeterWorkflow(ctx workflow.Context, params GCPAccessContextManagerServicePerimeterWorkflowParams) (*GCPAccessContextManagerServicePerimeterWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPAccessContextManagerServicePerimeterWorkflow")

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

	var result IngestServicePerimetersResult
	err := workflow.ExecuteActivity(activityCtx, IngestServicePerimetersActivity, IngestServicePerimetersParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest service perimeters", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPAccessContextManagerServicePerimeterWorkflow",
		"perimeterCount", result.PerimeterCount,
	)

	return &GCPAccessContextManagerServicePerimeterWorkflowResult{
		PerimeterCount: result.PerimeterCount,
	}, nil
}
