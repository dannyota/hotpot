package ubuntu

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
)

// UbuntuPackagesWorkflowResult contains the result of the Ubuntu packages workflow.
type UbuntuPackagesWorkflowResult struct {
	PackageCount   int
	DurationMillis int64
}

// UbuntuPackagesWorkflow ingests Ubuntu package indexes, one activity per feed.
func UbuntuPackagesWorkflow(ctx workflow.Context) (*UbuntuPackagesWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting UbuntuPackagesWorkflow", "feedCount", len(Feeds))

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

	for i, feed := range Feeds {
		label := feed.Release + "/" + feed.Component
		logger.Info("Ingesting Ubuntu feed", "feed", label, "progress", fmt.Sprintf("%d/%d", i+1, len(Feeds)))

		var result IngestUbuntuFeedResult
		err := workflow.ExecuteActivity(activityCtx, IngestUbuntuFeedActivity, IngestUbuntuFeedInput{
			Release:   feed.Release,
			Component: feed.Component,
		}).Get(ctx, &result)
		if err != nil {
			logger.Error("Failed to ingest Ubuntu feed", "feed", label, "error", err)
			failures++
			continue
		}

		totalPackages += result.PackageCount
		logger.Info("Completed Ubuntu feed", "feed", label, "packageCount", result.PackageCount)
	}

	if failures == len(Feeds) {
		return nil, temporalerr.PropagateNonRetryable(fmt.Errorf("all %d Ubuntu feeds failed", failures))
	}

	logger.Info("Completed UbuntuPackagesWorkflow",
		"packageCount", totalPackages,
		"failures", failures,
	)

	return &UbuntuPackagesWorkflowResult{
		PackageCount:   totalPackages,
		DurationMillis: workflow.Now(ctx).Sub(start).Milliseconds(),
	}, nil
}
