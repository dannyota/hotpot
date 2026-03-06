package software

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// NormalizeSoftwareWorkflowResult holds the final result.
type NormalizeSoftwareWorkflowResult struct {
	NormalizeResults []NormalizeProviderResult
	MergeResult      MergeSoftwareResult
}

// NormalizeSoftwareWorkflow runs the two-phase normalize+merge pipeline.
func NormalizeSoftwareWorkflow(ctx workflow.Context) (*NormalizeSoftwareWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting NormalizeSoftwareWorkflow")

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

	// Phase 1: Normalize all providers in parallel.
	providerKeys := []string{"s1", "meec"}
	futures := make([]workflow.Future, len(providerKeys))
	for i, key := range providerKeys {
		futures[i] = workflow.ExecuteActivity(activityCtx, NormalizeSoftwareProviderActivity,
			NormalizeProviderParams{ProviderKey: key})
	}

	// Wait for all normalizations.
	result := &NormalizeSoftwareWorkflowResult{
		NormalizeResults: make([]NormalizeProviderResult, 0, len(providerKeys)),
	}
	var errs []error
	for i, f := range futures {
		var nr NormalizeProviderResult
		if err := f.Get(ctx, &nr); err != nil {
			logger.Error("Failed to normalize provider", "provider", providerKeys[i], "error", err)
			errs = append(errs, err)
		} else {
			result.NormalizeResults = append(result.NormalizeResults, nr)
		}
	}

	// Phase 2: Merge normalized rows into final software.
	var mergeResult MergeSoftwareResult
	if err := workflow.ExecuteActivity(activityCtx, MergeSoftwareActivity).Get(ctx, &mergeResult); err != nil {
		logger.Error("Failed to merge software", "error", err)
		return result, err
	}
	result.MergeResult = mergeResult

	logger.Info("Completed NormalizeSoftwareWorkflow",
		"created", mergeResult.Created,
		"updated", mergeResult.Updated,
		"deleted", mergeResult.Deleted)

	if len(errs) > 0 {
		logger.Warn("NormalizeSoftwareWorkflow completed with provider errors", "errorCount", len(errs))
	}

	return result, nil
}
