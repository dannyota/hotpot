package job

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// JenkinsJobWorkflowResult contains the result of the job workflow.
type JenkinsJobWorkflowResult struct {
	JobCount       int
	BuildCount     int
	RepoCount      int
	DurationMillis int64
}

// JenkinsJobWorkflow ingests Jenkins jobs, builds, and repos.
func JenkinsJobWorkflow(ctx workflow.Context) (*JenkinsJobWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting JenkinsJobWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestJenkinsJobsResult
	err := workflow.ExecuteActivity(activityCtx, IngestJenkinsJobsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest jobs", "error", err)
		return nil, err
	}

	logger.Info("Completed JenkinsJobWorkflow",
		"jobCount", result.JobCount,
		"buildCount", result.BuildCount,
		"repoCount", result.RepoCount,
	)

	return &JenkinsJobWorkflowResult{
		JobCount:       result.JobCount,
		BuildCount:     result.BuildCount,
		RepoCount:      result.RepoCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
