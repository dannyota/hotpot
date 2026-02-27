package jenkins

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest"
)

// JenkinsInventoryWorkflowResult contains the result of Jenkins inventory collection.
type JenkinsInventoryWorkflowResult struct {
	JobCount   int
	BuildCount int
	RepoCount  int
}

// aggregateFunc is the function signature for merging a service result into the provider result.
type aggregateFunc = func(*JenkinsInventoryWorkflowResult, any)

// JenkinsInventoryWorkflow orchestrates Jenkins inventory collection.
func JenkinsInventoryWorkflow(ctx workflow.Context) (*JenkinsInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting JenkinsInventoryWorkflow")

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 60 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, childOpts)

	result := &JenkinsInventoryWorkflowResult{}

	for _, svc := range ingest.Services("jenkins") {
		res := svc.NewResult()
		err := workflow.ExecuteChildWorkflow(ctx, svc.Workflow).Get(ctx, res)
		if err != nil {
			logger.Error("Failed ingestion", "service", svc.Name, "error", err)
		} else {
			svc.Aggregate.(aggregateFunc)(result, res)
		}
	}

	logger.Info("Completed JenkinsInventoryWorkflow",
		"jobs", result.JobCount,
		"builds", result.BuildCount,
		"repos", result.RepoCount,
	)

	return result, nil
}
