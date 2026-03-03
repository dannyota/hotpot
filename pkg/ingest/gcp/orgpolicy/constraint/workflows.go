package constraint

import (
	"time"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type GCPOrgPolicyConstraintWorkflowParams struct {
}

type GCPOrgPolicyConstraintWorkflowResult struct {
	ConstraintCount int
}

func GCPOrgPolicyConstraintWorkflow(ctx workflow.Context, params GCPOrgPolicyConstraintWorkflowParams) (*GCPOrgPolicyConstraintWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPOrgPolicyConstraintWorkflow")

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

	var result IngestConstraintsResult
	err := workflow.ExecuteActivity(activityCtx, IngestConstraintsActivity, IngestConstraintsParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest org policy constraints", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed GCPOrgPolicyConstraintWorkflow",
		"constraintCount", result.ConstraintCount,
	)

	return &GCPOrgPolicyConstraintWorkflowResult{
		ConstraintCount: result.ConstraintCount,
	}, nil
}
