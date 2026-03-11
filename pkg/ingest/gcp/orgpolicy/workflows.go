package orgpolicy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
	"danny.vn/hotpot/pkg/ingest/gcp/orgpolicy/constraint"
	"danny.vn/hotpot/pkg/ingest/gcp/orgpolicy/customconstraint"
	"danny.vn/hotpot/pkg/ingest/gcp/orgpolicy/policy"
)

// GCPOrgPolicyWorkflowParams contains parameters for the org policy workflow.
type GCPOrgPolicyWorkflowParams struct {
}

// GCPOrgPolicyWorkflowResult contains the result of the org policy workflow.
type GCPOrgPolicyWorkflowResult struct {
	ConstraintCount       int
	CustomConstraintCount int
	PolicyCount           int
}

// GCPOrgPolicyWorkflow ingests all Organization Policy resources.
// Constraints and custom constraints run in parallel, then policies run after (they reference constraints).
func GCPOrgPolicyWorkflow(ctx workflow.Context, params GCPOrgPolicyWorkflowParams) (*GCPOrgPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPOrgPolicyWorkflow")

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 60 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPOrgPolicyWorkflowResult{}

	// Phase 1: Ingest constraints and custom constraints in parallel (independent)
	constraintFuture := workflow.ExecuteChildWorkflow(childCtx, constraint.GCPOrgPolicyConstraintWorkflow,
		constraint.GCPOrgPolicyConstraintWorkflowParams{})

	customConstraintFuture := workflow.ExecuteChildWorkflow(childCtx, customconstraint.GCPOrgPolicyCustomConstraintWorkflow,
		customconstraint.GCPOrgPolicyCustomConstraintWorkflowParams{})

	// Collect constraint results
	var constraintResult constraint.GCPOrgPolicyConstraintWorkflowResult
	if err := constraintFuture.Get(ctx, &constraintResult); err != nil {
		logger.Error("Failed to ingest org policy constraints", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}
	result.ConstraintCount = constraintResult.ConstraintCount

	// Collect custom constraint results
	var customConstraintResult customconstraint.GCPOrgPolicyCustomConstraintWorkflowResult
	if err := customConstraintFuture.Get(ctx, &customConstraintResult); err != nil {
		logger.Error("Failed to ingest org policy custom constraints", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}
	result.CustomConstraintCount = customConstraintResult.CustomConstraintCount

	// Phase 2: Ingest policies (depends on constraints being in DB)
	var policyResult policy.GCPOrgPolicyPolicyWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, policy.GCPOrgPolicyPolicyWorkflow,
		policy.GCPOrgPolicyPolicyWorkflowParams{}).Get(ctx, &policyResult)
	if err != nil {
		logger.Error("Failed to ingest org policies", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}
	result.PolicyCount = policyResult.PolicyCount

	logger.Info("Completed GCPOrgPolicyWorkflow",
		"constraintCount", result.ConstraintCount,
		"customConstraintCount", result.CustomConstraintCount,
		"policyCount", result.PolicyCount,
	)

	return result, nil
}
