package iampolicysearch

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPCloudAssetIAMPolicySearchWorkflowParams contains parameters for the IAM policy search workflow.
type GCPCloudAssetIAMPolicySearchWorkflowParams struct {
}

// GCPCloudAssetIAMPolicySearchWorkflowResult contains the result of the IAM policy search workflow.
type GCPCloudAssetIAMPolicySearchWorkflowResult struct {
	PolicyCount int
}

// GCPCloudAssetIAMPolicySearchWorkflow ingests IAM policy search results.
func GCPCloudAssetIAMPolicySearchWorkflow(ctx workflow.Context, params GCPCloudAssetIAMPolicySearchWorkflowParams) (*GCPCloudAssetIAMPolicySearchWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPCloudAssetIAMPolicySearchWorkflow")

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

	var result IngestIAMPolicySearchResult
	err := workflow.ExecuteActivity(activityCtx, IngestIAMPolicySearchActivity, IngestIAMPolicySearchParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest IAM policy search results", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPCloudAssetIAMPolicySearchWorkflow",
		"policyCount", result.PolicyCount,
	)

	return &GCPCloudAssetIAMPolicySearchWorkflowResult{
		PolicyCount: result.PolicyCount,
	}, nil
}
