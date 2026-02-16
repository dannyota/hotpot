package iap

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/iap/iampolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iap/settings"
)

// GCPIAPWorkflowParams contains parameters for the IAP workflow.
type GCPIAPWorkflowParams struct {
	ProjectID string
}

// GCPIAPWorkflowResult contains the result of the IAP workflow.
type GCPIAPWorkflowResult struct {
	SettingsCount int
	PolicyCount   int
}

// GCPIAPWorkflow ingests all Identity-Aware Proxy resources.
// Executes settings workflow first, then IAM policy.
func GCPIAPWorkflow(ctx workflow.Context, params GCPIAPWorkflowParams) (*GCPIAPWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPIAPWorkflow", "projectID", params.ProjectID)

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 60 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPIAPWorkflowResult{}

	// Phase 1: Ingest IAP settings
	var settingsResult settings.GCPIAPSettingsWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, settings.GCPIAPSettingsWorkflow,
		settings.GCPIAPSettingsWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &settingsResult)
	if err != nil {
		logger.Error("Failed to ingest IAP settings", "error", err)
		return nil, err
	}
	result.SettingsCount = settingsResult.SettingsCount

	// Phase 2: Ingest IAP IAM policy
	var policyResult iampolicy.GCPIAPIAMPolicyWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, iampolicy.GCPIAPIAMPolicyWorkflow,
		iampolicy.GCPIAPIAMPolicyWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &policyResult)
	if err != nil {
		logger.Error("Failed to ingest IAP IAM policy", "error", err)
		return nil, err
	}
	result.PolicyCount = policyResult.PolicyCount

	logger.Info("Completed GCPIAPWorkflow",
		"projectID", params.ProjectID,
		"settingsCount", result.SettingsCount,
		"policyCount", result.PolicyCount,
	)

	return result, nil
}
