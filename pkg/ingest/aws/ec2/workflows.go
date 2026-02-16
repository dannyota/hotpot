package ec2

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/aws/ec2/instance"
)

// AWSEC2WorkflowParams contains parameters for the EC2 workflow.
type AWSEC2WorkflowParams struct {
	Region string
}

// AWSEC2WorkflowResult contains the result of the EC2 workflow.
type AWSEC2WorkflowResult struct {
	Region        string
	InstanceCount int
}

// AWSEC2Workflow ingests all EC2 resources for a single region.
func AWSEC2Workflow(ctx workflow.Context, params AWSEC2WorkflowParams) (*AWSEC2WorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting AWSEC2Workflow", "region", params.Region)

	// Child workflow options
	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, childOpts)

	result := &AWSEC2WorkflowResult{Region: params.Region}

	// Execute EC2 Instance workflow
	var instanceResult instance.AWSEC2InstanceWorkflowResult
	err := workflow.ExecuteChildWorkflow(ctx, instance.AWSEC2InstanceWorkflow, instance.AWSEC2InstanceWorkflowParams{
		Region: params.Region,
	}).Get(ctx, &instanceResult)

	if err != nil {
		logger.Error("Failed to execute AWSEC2InstanceWorkflow", "region", params.Region, "error", err)
		return nil, err
	}

	result.InstanceCount = instanceResult.InstanceCount

	logger.Info("Completed AWSEC2Workflow",
		"region", params.Region,
		"instanceCount", result.InstanceCount,
	)

	return result, nil
}
