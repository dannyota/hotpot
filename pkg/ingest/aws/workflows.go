package aws

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/aws/ec2"
)

// AWSInventoryWorkflowParams contains parameters for the AWS inventory workflow.
type AWSInventoryWorkflowParams struct{}

// AWSInventoryWorkflowResult contains the result of the AWS inventory workflow.
type AWSInventoryWorkflowResult struct {
	RegionResults  []RegionResult
	TotalInstances int
}

// RegionResult contains the ingestion result for a single region.
type RegionResult struct {
	Region        string
	InstanceCount int
	Error         string
}

// AWSInventoryWorkflow ingests all AWS resources across all enabled regions.
// It first discovers regions, then orchestrates per-region child workflows.
func AWSInventoryWorkflow(ctx workflow.Context, _ AWSInventoryWorkflowParams) (*AWSInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting AWSInventoryWorkflow")

	// Activity options for region discovery
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	// Discover regions
	var discoverResult DiscoverRegionsResult
	err := workflow.ExecuteActivity(activityCtx, DiscoverRegionsActivity, DiscoverRegionsParams{}).
		Get(ctx, &discoverResult)
	if err != nil {
		logger.Error("Failed to discover regions", "error", err)
		return nil, err
	}

	logger.Info("Discovered regions", "count", len(discoverResult.Regions))

	// Child workflow options
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

	result := &AWSInventoryWorkflowResult{
		RegionResults: make([]RegionResult, 0, len(discoverResult.Regions)),
	}

	// Process each region
	for _, region := range discoverResult.Regions {
		regionResult := RegionResult{Region: region}

		// Execute AWSEC2Workflow for this region
		var ec2Result ec2.AWSEC2WorkflowResult
		err := workflow.ExecuteChildWorkflow(ctx, ec2.AWSEC2Workflow, ec2.AWSEC2WorkflowParams{
			Region: region,
		}).Get(ctx, &ec2Result)

		if err != nil {
			logger.Error("Failed to execute AWSEC2Workflow", "region", region, "error", err)
			regionResult.Error = err.Error()
		} else {
			regionResult.InstanceCount = ec2Result.InstanceCount
			result.TotalInstances += ec2Result.InstanceCount
		}

		result.RegionResults = append(result.RegionResults, regionResult)
	}

	logger.Info("Completed AWSInventoryWorkflow",
		"regionCount", len(discoverResult.Regions),
		"totalInstances", result.TotalInstances,
	)

	return result, nil
}
