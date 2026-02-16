package alertpolicy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPMonitoringAlertPolicyWorkflowParams contains parameters for the alert policy workflow.
type GCPMonitoringAlertPolicyWorkflowParams struct {
	ProjectID string
}

// GCPMonitoringAlertPolicyWorkflowResult contains the result of the alert policy workflow.
type GCPMonitoringAlertPolicyWorkflowResult struct {
	ProjectID        string
	AlertPolicyCount int
	DurationMillis   int64
}

// GCPMonitoringAlertPolicyWorkflow ingests Monitoring alert policies for a single project.
func GCPMonitoringAlertPolicyWorkflow(ctx workflow.Context, params GCPMonitoringAlertPolicyWorkflowParams) (*GCPMonitoringAlertPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPMonitoringAlertPolicyWorkflow", "projectID", params.ProjectID)

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

	var result IngestAlertPoliciesResult
	err := workflow.ExecuteActivity(activityCtx, IngestAlertPoliciesActivity, IngestAlertPoliciesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest alert policies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPMonitoringAlertPolicyWorkflow",
		"projectID", params.ProjectID,
		"alertPolicyCount", result.AlertPolicyCount,
	)

	return &GCPMonitoringAlertPolicyWorkflowResult{
		ProjectID:        result.ProjectID,
		AlertPolicyCount: result.AlertPolicyCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
