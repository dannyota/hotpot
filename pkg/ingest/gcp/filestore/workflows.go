package filestore

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/filestore/instance"
)

// GCPFilestoreWorkflowParams contains parameters for the Filestore workflow.
type GCPFilestoreWorkflowParams struct {
	ProjectID string
}

// GCPFilestoreWorkflowResult contains the result of the Filestore workflow.
type GCPFilestoreWorkflowResult struct {
	ProjectID     string
	InstanceCount int
}

// GCPFilestoreWorkflow ingests all GCP Filestore resources for a single project.
func GCPFilestoreWorkflow(ctx workflow.Context, params GCPFilestoreWorkflowParams) (*GCPFilestoreWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPFilestoreWorkflow", "projectID", params.ProjectID)

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

	result := &GCPFilestoreWorkflowResult{
		ProjectID: params.ProjectID,
	}

	var instanceResult instance.GCPFilestoreInstanceWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, instance.GCPFilestoreInstanceWorkflow,
		instance.GCPFilestoreInstanceWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &instanceResult)
	if err != nil {
		logger.Error("Failed to ingest Filestore instances", "error", err)
		return nil, err
	}
	result.InstanceCount = instanceResult.InstanceCount

	logger.Info("Completed GCPFilestoreWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
	)

	return result, nil
}
