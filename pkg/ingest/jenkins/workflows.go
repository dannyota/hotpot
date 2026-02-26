package jenkins

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/jenkins/job"
)

// JenkinsInventoryWorkflowResult contains the result of Jenkins inventory collection.
type JenkinsInventoryWorkflowResult struct {
	JobCount   int
	BuildCount int
	RepoCount  int
}

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

	var jobResult job.JenkinsJobWorkflowResult
	err := workflow.ExecuteChildWorkflow(ctx, job.JenkinsJobWorkflow).Get(ctx, &jobResult)
	if err != nil {
		logger.Error("Failed to execute JenkinsJobWorkflow", "error", err)
	} else {
		result.JobCount = jobResult.JobCount
		result.BuildCount = jobResult.BuildCount
		result.RepoCount = jobResult.RepoCount
	}

	logger.Info("Completed JenkinsInventoryWorkflow",
		"jobs", result.JobCount,
		"builds", result.BuildCount,
		"repos", result.RepoCount,
	)

	return result, nil
}
