package customconstraint

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type GCPOrgPolicyCustomConstraintWorkflowParams struct {
}

type GCPOrgPolicyCustomConstraintWorkflowResult struct {
	CustomConstraintCount int
}

func GCPOrgPolicyCustomConstraintWorkflow(ctx workflow.Context, params GCPOrgPolicyCustomConstraintWorkflowParams) (*GCPOrgPolicyCustomConstraintWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPOrgPolicyCustomConstraintWorkflow")

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

	var result IngestCustomConstraintsResult
	err := workflow.ExecuteActivity(activityCtx, IngestCustomConstraintsActivity, IngestCustomConstraintsParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest org policy custom constraints", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPOrgPolicyCustomConstraintWorkflow",
		"customConstraintCount", result.CustomConstraintCount,
	)

	return &GCPOrgPolicyCustomConstraintWorkflowResult{
		CustomConstraintCount: result.CustomConstraintCount,
	}, nil
}
