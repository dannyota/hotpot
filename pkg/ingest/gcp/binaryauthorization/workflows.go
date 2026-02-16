package binaryauthorization

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/binaryauthorization/attestor"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/binaryauthorization/policy"
)

// GCPBinaryAuthorizationWorkflowParams contains parameters for the Binary Authorization workflow.
type GCPBinaryAuthorizationWorkflowParams struct {
	ProjectID string
}

// GCPBinaryAuthorizationWorkflowResult contains the result of the Binary Authorization workflow.
type GCPBinaryAuthorizationWorkflowResult struct {
	ProjectID     string
	PolicyCount   int
	AttestorCount int
}

// GCPBinaryAuthorizationWorkflow ingests all Binary Authorization resources for a single project.
// Executes policy workflow first, then attestors.
func GCPBinaryAuthorizationWorkflow(ctx workflow.Context, params GCPBinaryAuthorizationWorkflowParams) (*GCPBinaryAuthorizationWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPBinaryAuthorizationWorkflow", "projectID", params.ProjectID)

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPBinaryAuthorizationWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Phase 1: Ingest policy
	var policyResult policy.GCPBinaryAuthorizationPolicyWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, policy.GCPBinaryAuthorizationPolicyWorkflow,
		policy.GCPBinaryAuthorizationPolicyWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &policyResult)
	if err != nil {
		logger.Error("Failed to ingest binary authorization policy", "error", err)
		return nil, err
	}
	result.PolicyCount = policyResult.PolicyCount

	// Phase 2: Ingest attestors
	var attestorResult attestor.GCPBinaryAuthorizationAttestorWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, attestor.GCPBinaryAuthorizationAttestorWorkflow,
		attestor.GCPBinaryAuthorizationAttestorWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &attestorResult)
	if err != nil {
		logger.Error("Failed to ingest binary authorization attestors", "error", err)
		return nil, err
	}
	result.AttestorCount = attestorResult.AttestorCount

	logger.Info("Completed GCPBinaryAuthorizationWorkflow",
		"projectID", params.ProjectID,
		"policyCount", result.PolicyCount,
		"attestorCount", result.AttestorCount,
	)

	return result, nil
}
