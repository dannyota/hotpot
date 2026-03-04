package rpm

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
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
		StartToCloseTimeout: 30 * time.Minute,
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
	totalPackages := 0
	failures := 0

	for i, repo := range Repos {
		logger.Info("Ingesting RPM repo", "repo", repo.Name, "progress", fmt.Sprintf("%d/%d", i+1, len(Repos)))

		var result IngestRPMRepoResult
		err := workflow.ExecuteActivity(activityCtx, IngestRPMRepoActivity, IngestRPMRepoInput{
			RepoName: repo.Name,
		}).Get(ctx, &result)
		if err != nil {
			logger.Error("Failed to ingest RPM repo", "repo", repo.Name, "error", err)
			failures++
			continue
		}

		totalPackages += result.PackageCount
		logger.Info("Completed RPM repo", "repo", repo.Name, "packageCount", result.PackageCount)
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
