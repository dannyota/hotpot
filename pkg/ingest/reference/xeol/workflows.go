package xeol

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
)

// XeolWorkflowResult contains the result of the xeol workflow.
type XeolWorkflowResult struct {
	ProductCount   int
	CycleCount     int
	PurlCount      int
	VulnCount      int
	DurationMillis int64
}

// XeolWorkflow ingests the xeol EOL database.
func XeolWorkflow(ctx workflow.Context) (*XeolWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting XeolWorkflow")

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

	var result IngestXeolResult
	err := workflow.ExecuteActivity(activityCtx, IngestXeolActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest xeol data", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed XeolWorkflow",
		"productCount", result.ProductCount,
		"cycleCount", result.CycleCount,
		"purlCount", result.PurlCount,
		"vulnCount", result.VulnCount,
	)

	return &XeolWorkflowResult{
		ProductCount:   result.ProductCount,
		CycleCount:     result.CycleCount,
		PurlCount:      result.PurlCount,
		VulnCount:      result.VulnCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
