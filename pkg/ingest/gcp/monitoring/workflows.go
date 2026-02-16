package monitoring

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/monitoring/alertpolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/monitoring/uptimecheck"
)

// GCPMonitoringWorkflowParams contains parameters for the monitoring workflow.
type GCPMonitoringWorkflowParams struct {
	ProjectID string
}

// GCPMonitoringWorkflowResult contains the result of the monitoring workflow.
type GCPMonitoringWorkflowResult struct {
	ProjectID        string
	AlertPolicyCount int
	UptimeCheckCount int
}

// GCPMonitoringWorkflow ingests all GCP Monitoring resources for a single project.
// Alert policies and uptime checks are independent and run in parallel.
func GCPMonitoringWorkflow(ctx workflow.Context, params GCPMonitoringWorkflowParams) (*GCPMonitoringWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPMonitoringWorkflow", "projectID", params.ProjectID)

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPMonitoringWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Execute alert policy and uptime check workflows in parallel
	alertPolicyFuture := workflow.ExecuteChildWorkflow(childCtx, alertpolicy.GCPMonitoringAlertPolicyWorkflow,
		alertpolicy.GCPMonitoringAlertPolicyWorkflowParams{ProjectID: params.ProjectID})

	uptimeCheckFuture := workflow.ExecuteChildWorkflow(childCtx, uptimecheck.GCPMonitoringUptimeCheckWorkflow,
		uptimecheck.GCPMonitoringUptimeCheckWorkflowParams{ProjectID: params.ProjectID})

	// Wait for alert policy result
	var alertPolicyResult alertpolicy.GCPMonitoringAlertPolicyWorkflowResult
	if err := alertPolicyFuture.Get(ctx, &alertPolicyResult); err != nil {
		logger.Error("Failed to ingest alert policies", "error", err)
		return nil, err
	}
	result.AlertPolicyCount = alertPolicyResult.AlertPolicyCount

	// Wait for uptime check result
	var uptimeCheckResult uptimecheck.GCPMonitoringUptimeCheckWorkflowResult
	if err := uptimeCheckFuture.Get(ctx, &uptimeCheckResult); err != nil {
		logger.Error("Failed to ingest uptime check configs", "error", err)
		return nil, err
	}
	result.UptimeCheckCount = uptimeCheckResult.UptimeCheckCount

	logger.Info("Completed GCPMonitoringWorkflow",
		"projectID", params.ProjectID,
		"alertPolicyCount", result.AlertPolicyCount,
		"uptimeCheckCount", result.UptimeCheckCount,
	)

	return result, nil
}
