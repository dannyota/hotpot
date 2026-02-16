package managedzone

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPDNSManagedZoneWorkflowParams contains parameters for the managed zone workflow.
type GCPDNSManagedZoneWorkflowParams struct {
	ProjectID string
}

// GCPDNSManagedZoneWorkflowResult contains the result of the managed zone workflow.
type GCPDNSManagedZoneWorkflowResult struct {
	ProjectID        string
	ManagedZoneCount int
	DurationMillis   int64
}

// GCPDNSManagedZoneWorkflow ingests GCP DNS managed zones for a single project.
func GCPDNSManagedZoneWorkflow(ctx workflow.Context, params GCPDNSManagedZoneWorkflowParams) (*GCPDNSManagedZoneWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPDNSManagedZoneWorkflow", "projectID", params.ProjectID)

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
	var result IngestDNSManagedZonesResult
	err := workflow.ExecuteActivity(activityCtx, IngestDNSManagedZonesActivity, IngestDNSManagedZonesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest managed zones", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPDNSManagedZoneWorkflow",
		"projectID", params.ProjectID,
		"managedZoneCount", result.ManagedZoneCount,
	)

	return &GCPDNSManagedZoneWorkflowResult{
		ProjectID:        result.ProjectID,
		ManagedZoneCount: result.ManagedZoneCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
