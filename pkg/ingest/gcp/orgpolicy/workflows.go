package orgpolicy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/orgpolicy/constraint"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/orgpolicy/customconstraint"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/orgpolicy/policy"
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
// Executes constraint and policy workflows sequentially.
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

	// Phase 1: Ingest constraints
	var constraintResult constraint.GCPOrgPolicyConstraintWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, constraint.GCPOrgPolicyConstraintWorkflow,
		constraint.GCPOrgPolicyConstraintWorkflowParams{}).Get(ctx, &constraintResult)
	if err != nil {
		logger.Error("Failed to ingest org policy constraints", "error", err)
		return nil, err
	}
	result.ConstraintCount = constraintResult.ConstraintCount

	// Phase 2: Ingest custom constraints
	var customConstraintResult customconstraint.GCPOrgPolicyCustomConstraintWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, customconstraint.GCPOrgPolicyCustomConstraintWorkflow,
		customconstraint.GCPOrgPolicyCustomConstraintWorkflowParams{}).Get(ctx, &customConstraintResult)
	if err != nil {
		logger.Error("Failed to ingest org policy custom constraints", "error", err)
		return nil, err
	}
	result.CustomConstraintCount = customConstraintResult.CustomConstraintCount

	// Phase 3: Ingest policies
	var policyResult policy.GCPOrgPolicyPolicyWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, policy.GCPOrgPolicyPolicyWorkflow,
		policy.GCPOrgPolicyPolicyWorkflowParams{}).Get(ctx, &policyResult)
	if err != nil {
		logger.Error("Failed to ingest org policies", "error", err)
		return nil, err
	}
	result.PolicyCount = policyResult.PolicyCount

	logger.Info("Completed GCPOrgPolicyWorkflow",
		"constraintCount", result.ConstraintCount,
		"customConstraintCount", result.CustomConstraintCount,
		"policyCount", result.PolicyCount,
	)

	return result, nil
}
