package glbpackage

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeGLBGlobalPackageWorkflowParams contains parameters for the package workflow.
type GreenNodeGLBGlobalPackageWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeGLBGlobalPackageWorkflowResult contains the result of the package workflow.
type GreenNodeGLBGlobalPackageWorkflowResult struct {
	PackageCount   int
	DurationMillis int64
}

// GreenNodeGLBGlobalPackageWorkflow ingests GreenNode global packages.
func GreenNodeGLBGlobalPackageWorkflow(ctx workflow.Context, params GreenNodeGLBGlobalPackageWorkflowParams) (*GreenNodeGLBGlobalPackageWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeGLBGlobalPackageWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestGLBGlobalPackagesResult
	err := workflow.ExecuteActivity(activityCtx, IngestGLBGlobalPackagesActivity, IngestGLBGlobalPackagesParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest global packages", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeGLBGlobalPackageWorkflow", "packageCount", result.PackageCount)

	return &GreenNodeGLBGlobalPackageWorkflowResult{
		PackageCount:   result.PackageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
