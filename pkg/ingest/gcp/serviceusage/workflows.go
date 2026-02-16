package serviceusage

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/serviceusage/enabledservice"
)

// GCPServiceUsageWorkflowParams contains parameters for the Service Usage workflow.
type GCPServiceUsageWorkflowParams struct {
	ProjectID string
}

// GCPServiceUsageWorkflowResult contains the result of the Service Usage workflow.
type GCPServiceUsageWorkflowResult struct {
	ProjectID    string
	ServiceCount int
}

// GCPServiceUsageWorkflow ingests all GCP Service Usage resources for a single project.
func GCPServiceUsageWorkflow(ctx workflow.Context, params GCPServiceUsageWorkflowParams) (*GCPServiceUsageWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPServiceUsageWorkflow", "projectID", params.ProjectID)

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

	result := &GCPServiceUsageWorkflowResult{
		ProjectID: params.ProjectID,
	}

	var enabledServiceResult enabledservice.GCPServiceUsageEnabledServiceWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, enabledservice.GCPServiceUsageEnabledServiceWorkflow,
		enabledservice.GCPServiceUsageEnabledServiceWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &enabledServiceResult)
	if err != nil {
		logger.Error("Failed to ingest enabled services", "error", err)
		return nil, err
	}
	result.ServiceCount = enabledServiceResult.ServiceCount

	logger.Info("Completed GCPServiceUsageWorkflow",
		"projectID", params.ProjectID,
		"serviceCount", result.ServiceCount,
	)

	return result, nil
}
