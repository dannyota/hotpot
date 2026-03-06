package aws

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/ingest"
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

// aggregateFunc is the function signature for merging a service result into the provider-level results.
type aggregateFunc = func(*AWSInventoryWorkflowResult, *RegionResult, any)

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

	services := ingest.Services("aws")

	// Process each region
	for _, region := range discoverResult.Regions {
		regionResult := RegionResult{Region: region}

		for _, svc := range services {
			res := svc.NewResult()
			err := workflow.ExecuteChildWorkflow(ctx, svc.Workflow,
				svc.NewParams("", region)).Get(ctx, res)
			if err != nil {
				logger.Error("Failed ingestion", "service", svc.Name, "region", region, "error", err)
				appendError(&regionResult, err)
			} else {
				svc.Aggregate.(aggregateFunc)(result, &regionResult, res)
			}
		}

		result.RegionResults = append(result.RegionResults, regionResult)
	}

	logger.Info("Completed AWSInventoryWorkflow",
		"regionCount", len(discoverResult.Regions),
		"totalInstances", result.TotalInstances,
	)

	return result, nil
}

func appendError(rr *RegionResult, err error) {
	if rr.Error == "" {
		rr.Error = err.Error()
	} else {
		rr.Error += "; " + err.Error()
	}
}
