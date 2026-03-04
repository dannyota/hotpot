package k8snode

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// NormalizeK8sNodesWorkflowResult holds the final result.
type NormalizeK8sNodesWorkflowResult struct {
	NormalizeResults []NormalizeProviderResult
	MergeResult      MergeK8sNodesResult
}

// NormalizeK8sNodesWorkflow runs the two-phase normalize+merge pipeline.
func NormalizeK8sNodesWorkflow(ctx workflow.Context) (*NormalizeK8sNodesWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting NormalizeK8sNodesWorkflow")

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
	providerKeys := []string{"gcp"}
	futures := make([]workflow.Future, len(providerKeys))
	for i, key := range providerKeys {
		futures[i] = workflow.ExecuteActivity(activityCtx, NormalizeProviderActivity,
			NormalizeProviderParams{ProviderKey: key})
	}

	// Wait for all normalizations.
	result := &NormalizeK8sNodesWorkflowResult{
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

	// Phase 2: Merge normalized rows into final k8s_nodes.
	var mergeResult MergeK8sNodesResult
	if err := workflow.ExecuteActivity(activityCtx, MergeK8sNodesActivity).Get(ctx, &mergeResult); err != nil {
		logger.Error("Failed to merge k8s nodes", "error", err)
		return result, err
	}
	result.MergeResult = mergeResult

	logger.Info("Completed NormalizeK8sNodesWorkflow",
		"created", mergeResult.Created,
		"updated", mergeResult.Updated,
		"deleted", mergeResult.Deleted)

	if len(errs) > 0 {
		logger.Warn("NormalizeK8sNodesWorkflow completed with provider errors", "errorCount", len(errs))
	}

	return result, nil
}
