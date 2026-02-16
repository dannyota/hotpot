package organization

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPResourceManagerOrganizationWorkflowParams contains parameters for the organization workflow.
type GCPResourceManagerOrganizationWorkflowParams struct {
	// Empty - discovers all accessible organizations
}

// GCPResourceManagerOrganizationWorkflowResult contains the result of the organization workflow.
type GCPResourceManagerOrganizationWorkflowResult struct {
	OrganizationCount int
	DurationMillis    int64
}

// GCPResourceManagerOrganizationWorkflow discovers all GCP organizations accessible by the service account.
func GCPResourceManagerOrganizationWorkflow(ctx workflow.Context, params GCPResourceManagerOrganizationWorkflowParams) (*GCPResourceManagerOrganizationWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPResourceManagerOrganizationWorkflow")

	// Activity options
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

	// Execute ingest activity
	var result IngestOrganizationsResult
	err := workflow.ExecuteActivity(activityCtx, IngestOrganizationsActivity, IngestOrganizationsParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to discover organizations", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPResourceManagerOrganizationWorkflow",
		"organizationCount", result.OrganizationCount,
	)

	return &GCPResourceManagerOrganizationWorkflowResult{
		OrganizationCount: result.OrganizationCount,
		DurationMillis:    result.DurationMillis,
	}, nil
}
