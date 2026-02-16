package cloudfunctions

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudfunctions/function"
)

// GCPCloudFunctionsWorkflowParams contains parameters for the Cloud Functions workflow.
type GCPCloudFunctionsWorkflowParams struct {
	ProjectID string
}

// GCPCloudFunctionsWorkflowResult contains the result of the Cloud Functions workflow.
type GCPCloudFunctionsWorkflowResult struct {
	ProjectID     string
	FunctionCount int
}

// GCPCloudFunctionsWorkflow ingests all GCP Cloud Functions resources for a single project.
func GCPCloudFunctionsWorkflow(ctx workflow.Context, params GCPCloudFunctionsWorkflowParams) (*GCPCloudFunctionsWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPCloudFunctionsWorkflow", "projectID", params.ProjectID)

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

	result := &GCPCloudFunctionsWorkflowResult{
		ProjectID: params.ProjectID,
	}

	var functionResult function.GCPCloudFunctionsFunctionWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, function.GCPCloudFunctionsFunctionWorkflow,
		function.GCPCloudFunctionsFunctionWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &functionResult)
	if err != nil {
		logger.Error("Failed to ingest Cloud Functions", "error", err)
		return nil, err
	}
	result.FunctionCount = functionResult.FunctionCount

	logger.Info("Completed GCPCloudFunctionsWorkflow",
		"projectID", params.ProjectID,
		"functionCount", result.FunctionCount,
	)

	return result, nil
}
