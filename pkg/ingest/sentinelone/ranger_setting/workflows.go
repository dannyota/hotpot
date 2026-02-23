package ranger_setting

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// S1RangerSettingWorkflowResult contains the result of the ranger setting workflow.
type S1RangerSettingWorkflowResult struct {
	SettingCount   int
	DurationMillis int64
}

// S1RangerSettingWorkflow ingests SentinelOne Ranger settings.
func S1RangerSettingWorkflow(ctx workflow.Context) (*S1RangerSettingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1RangerSettingWorkflow")

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

	var result IngestS1RangerSettingsResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1RangerSettingsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest ranger settings", "error", err)
		return nil, err
	}

	logger.Info("Completed S1RangerSettingWorkflow", "settingCount", result.SettingCount)

	return &S1RangerSettingWorkflowResult{
		SettingCount:   result.SettingCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
