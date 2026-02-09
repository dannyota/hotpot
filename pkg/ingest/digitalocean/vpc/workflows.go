package vpc

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DOVpcWorkflowResult contains the result of the VPC workflow.
type DOVpcWorkflowResult struct {
	VpcCount       int
	DurationMillis int64
}

// DOVpcWorkflow ingests DigitalOcean VPCs.
func DOVpcWorkflow(ctx workflow.Context) (*DOVpcWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DOVpcWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestDOVpcsResult
	err := workflow.ExecuteActivity(activityCtx, IngestDOVpcsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest VPCs", "error", err)
		return nil, err
	}

	logger.Info("Completed DOVpcWorkflow", "vpcCount", result.VpcCount)

	return &DOVpcWorkflowResult{
		VpcCount:       result.VpcCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
