package instance

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// AWSEC2InstanceWorkflowParams contains parameters for the instance workflow.
type AWSEC2InstanceWorkflowParams struct {
	Region string
}

// AWSEC2InstanceWorkflowResult contains the result of the instance workflow.
type AWSEC2InstanceWorkflowResult struct {
	Region         string
	InstanceCount  int
	DurationMillis int64
}

// AWSEC2InstanceWorkflow ingests AWS EC2 instances for a single region.
func AWSEC2InstanceWorkflow(ctx workflow.Context, params AWSEC2InstanceWorkflowParams) (*AWSEC2InstanceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting AWSEC2InstanceWorkflow", "region", params.Region)

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

	var result IngestEC2InstancesResult
	err := workflow.ExecuteActivity(activityCtx, IngestEC2InstancesActivity, IngestEC2InstancesParams{
		Region: params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest instances", "error", err)
		return nil, err
	}

	logger.Info("Completed AWSEC2InstanceWorkflow",
		"region", params.Region,
		"instanceCount", result.InstanceCount,
	)

	return &AWSEC2InstanceWorkflowResult{
		Region:         result.Region,
		InstanceCount:  result.InstanceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
