package enabledservice

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPServiceUsageEnabledServiceWorkflowParams contains parameters for the enabled service workflow.
type GCPServiceUsageEnabledServiceWorkflowParams struct {
	ProjectID string
}

// GCPServiceUsageEnabledServiceWorkflowResult contains the result of the enabled service workflow.
type GCPServiceUsageEnabledServiceWorkflowResult struct {
	ProjectID      string
	ServiceCount   int
	DurationMillis int64
}

// GCPServiceUsageEnabledServiceWorkflow ingests GCP enabled services for a single project.
func GCPServiceUsageEnabledServiceWorkflow(ctx workflow.Context, params GCPServiceUsageEnabledServiceWorkflowParams) (*GCPServiceUsageEnabledServiceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPServiceUsageEnabledServiceWorkflow", "projectID", params.ProjectID)

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

	var result IngestEnabledServicesResult
	err := workflow.ExecuteActivity(activityCtx, IngestEnabledServicesActivity, IngestEnabledServicesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest enabled services", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPServiceUsageEnabledServiceWorkflow",
		"projectID", params.ProjectID,
		"serviceCount", result.ServiceCount,
	)

	return &GCPServiceUsageEnabledServiceWorkflowResult{
		ProjectID:      result.ProjectID,
		ServiceCount:   result.ServiceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
