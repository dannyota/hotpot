package digitalocean

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/vpc"
)

// DOInventoryWorkflowResult contains the result of DigitalOcean inventory collection.
type DOInventoryWorkflowResult struct {
	VpcCount int
}

// DOInventoryWorkflow orchestrates DigitalOcean inventory collection.
func DOInventoryWorkflow(ctx workflow.Context) (*DOInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DOInventoryWorkflow")

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 60 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, childOpts)

	result := &DOInventoryWorkflowResult{}

	var vpcResult vpc.DOVpcWorkflowResult
	err := workflow.ExecuteChildWorkflow(ctx, vpc.DOVpcWorkflow).Get(ctx, &vpcResult)
	if err != nil {
		logger.Error("Failed to execute DOVpcWorkflow", "error", err)
	} else {
		result.VpcCount = vpcResult.VpcCount
	}

	logger.Info("Completed DOInventoryWorkflow", "vpcs", result.VpcCount)

	return result, nil
}
