package rpm

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
)

// RPMPackagesWorkflowResult contains the result of the RPM packages workflow.
type RPMPackagesWorkflowResult struct {
	PackageCount   int
	DurationMillis int64
}

// RPMPackagesWorkflow ingests RPM repository metadata, one activity per repo.
func RPMPackagesWorkflow(ctx workflow.Context) (*RPMPackagesWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting RPMPackagesWorkflow", "repoCount", len(Repos))

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 4 * time.Hour,
		HeartbeatTimeout:    5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	start := workflow.Now(ctx)

	// Fan out: launch all repo activities in parallel.
	type repoFuture struct {
		name   string
		future workflow.Future
	}
	futures := make([]repoFuture, len(Repos))
	for i, repo := range Repos {
		logger.Info("Launching RPM repo activity", "repo", repo.Name, "index", fmt.Sprintf("%d/%d", i+1, len(Repos)))
		futures[i] = repoFuture{
			name: repo.Name,
			future: workflow.ExecuteActivity(activityCtx, IngestRPMRepoActivity, IngestRPMRepoInput{
				RepoName: repo.Name,
			}),
		}
	}

	// Fan in: collect results.
	totalPackages := 0
	failures := 0
	for _, rf := range futures {
		var result IngestRPMRepoResult
		if err := rf.future.Get(ctx, &result); err != nil {
			logger.Error("Failed to ingest RPM repo", "repo", rf.name, "error", err)
			failures++
			continue
		}
		totalPackages += result.PackageCount
		logger.Info("Completed RPM repo", "repo", rf.name, "packageCount", result.PackageCount)
	}

	if failures == len(Repos) {
		return nil, temporalerr.PropagateNonRetryable(fmt.Errorf("all %d RPM repos failed", failures))
	}

	logger.Info("Completed RPMPackagesWorkflow",
		"packageCount", totalPackages,
		"failures", failures,
	)

	return &RPMPackagesWorkflowResult{
		PackageCount:   totalPackages,
		DurationMillis: workflow.Now(ctx).Sub(start).Milliseconds(),
	}, nil
}
