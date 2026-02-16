package redis

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/redis/instance"
)

// GCPRedisWorkflowParams contains parameters for the Redis workflow.
type GCPRedisWorkflowParams struct {
	ProjectID string
}

// GCPRedisWorkflowResult contains the result of the Redis workflow.
type GCPRedisWorkflowResult struct {
	ProjectID     string
	InstanceCount int
}

// GCPRedisWorkflow ingests all GCP Memorystore Redis resources for a single project.
func GCPRedisWorkflow(ctx workflow.Context, params GCPRedisWorkflowParams) (*GCPRedisWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPRedisWorkflow", "projectID", params.ProjectID)

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

	result := &GCPRedisWorkflowResult{
		ProjectID: params.ProjectID,
	}

	var instanceResult instance.GCPRedisInstanceWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, instance.GCPRedisInstanceWorkflow,
		instance.GCPRedisInstanceWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &instanceResult)
	if err != nil {
		logger.Error("Failed to ingest Redis instances", "error", err)
		return nil, err
	}
	result.InstanceCount = instanceResult.InstanceCount

	logger.Info("Completed GCPRedisWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
	)

	return result, nil
}
