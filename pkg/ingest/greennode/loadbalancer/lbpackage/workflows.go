package lbpackage

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeLoadBalancerPackageWorkflowParams contains parameters for the package workflow.
type GreenNodeLoadBalancerPackageWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeLoadBalancerPackageWorkflowResult contains the result of the package workflow.
type GreenNodeLoadBalancerPackageWorkflowResult struct {
	PackageCount   int
	DurationMillis int64
}

// GreenNodeLoadBalancerPackageWorkflow ingests GreenNode load balancer packages.
func GreenNodeLoadBalancerPackageWorkflow(ctx workflow.Context, params GreenNodeLoadBalancerPackageWorkflowParams) (*GreenNodeLoadBalancerPackageWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeLoadBalancerPackageWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestLoadBalancerPackagesResult
	err := workflow.ExecuteActivity(activityCtx, IngestLoadBalancerPackagesActivity, IngestLoadBalancerPackagesParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest packages", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeLoadBalancerPackageWorkflow", "packageCount", result.PackageCount)

	return &GreenNodeLoadBalancerPackageWorkflowResult{
		PackageCount:   result.PackageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
