package machine

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// NormalizeMachinesWorkflowResult holds the final result.
type NormalizeMachinesWorkflowResult struct {
	NormalizeResults []NormalizeProviderResult
	MergeResult      MergeMachinesResult
}

// NormalizeMachinesWorkflow runs the two-phase normalize+merge pipeline.
func NormalizeMachinesWorkflow(ctx workflow.Context) (*NormalizeMachinesWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting NormalizeMachinesWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	// Phase 1: Normalize all providers in parallel.
	providerKeys := []string{"s1", "meec", "gcp", "greennode"}
	futures := make([]workflow.Future, len(providerKeys))
	for i, key := range providerKeys {
		futures[i] = workflow.ExecuteActivity(activityCtx, NormalizeMachineProviderActivity,
			NormalizeProviderParams{ProviderKey: key})
	}

	// Wait for all normalizations.
	result := &NormalizeMachinesWorkflowResult{
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

	// Phase 2: Merge normalized rows into final machines.
	var mergeResult MergeMachinesResult
	if err := workflow.ExecuteActivity(activityCtx, MergeMachinesActivity).Get(ctx, &mergeResult); err != nil {
		logger.Error("Failed to merge machines", "error", err)
		return result, err
	}
	result.MergeResult = mergeResult

	logger.Info("Completed NormalizeMachinesWorkflow",
		"created", mergeResult.Created,
		"updated", mergeResult.Updated,
		"deleted", mergeResult.Deleted)

	if len(errs) > 0 {
		logger.Warn("NormalizeMachinesWorkflow completed with provider errors", "errorCount", len(errs))
	}

	return result, nil
}
