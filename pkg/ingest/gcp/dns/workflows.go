package dns

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/dns/managedzone"
)

// GCPDNSWorkflowParams contains parameters for the DNS workflow.
type GCPDNSWorkflowParams struct {
	ProjectID string
}

// GCPDNSWorkflowResult contains the result of the DNS workflow.
type GCPDNSWorkflowResult struct {
	ProjectID        string
	ManagedZoneCount int
}

// GCPDNSWorkflow ingests all GCP DNS resources for a single project.
func GCPDNSWorkflow(ctx workflow.Context, params GCPDNSWorkflowParams) (*GCPDNSWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPDNSWorkflow", "projectID", params.ProjectID)

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

	result := &GCPDNSWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Execute managed zone workflow
	var managedZoneResult managedzone.GCPDNSManagedZoneWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, managedzone.GCPDNSManagedZoneWorkflow,
		managedzone.GCPDNSManagedZoneWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &managedZoneResult)
	if err != nil {
		logger.Error("Failed to ingest managed zones", "error", err)
		return nil, err
	}
	result.ManagedZoneCount = managedZoneResult.ManagedZoneCount

	logger.Info("Completed GCPDNSWorkflow",
		"projectID", params.ProjectID,
		"managedZoneCount", result.ManagedZoneCount,
	)

	return result, nil
}
