package accesscontextmanager

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/accesscontextmanager/accesslevel"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/accesscontextmanager/accesspolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/accesscontextmanager/serviceperimeter"
)

// GCPAccessContextManagerWorkflowParams contains parameters for the Access Context Manager workflow.
type GCPAccessContextManagerWorkflowParams struct {
}

// GCPAccessContextManagerWorkflowResult contains the result of the Access Context Manager workflow.
type GCPAccessContextManagerWorkflowResult struct {
	PolicyCount    int
	LevelCount     int
	PerimeterCount int
}

// GCPAccessContextManagerWorkflow ingests all Access Context Manager resources.
// Executes access policy workflow first, then access levels and service perimeters in parallel
// (both depend on policies being in DB).
func GCPAccessContextManagerWorkflow(ctx workflow.Context, params GCPAccessContextManagerWorkflowParams) (*GCPAccessContextManagerWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPAccessContextManagerWorkflow")

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

	result := &GCPAccessContextManagerWorkflowResult{}

	// Phase 1: Ingest access policies first (levels and perimeters reference policies)
	var policyResult accesspolicy.GCPAccessContextManagerAccessPolicyWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, accesspolicy.GCPAccessContextManagerAccessPolicyWorkflow,
		accesspolicy.GCPAccessContextManagerAccessPolicyWorkflowParams{}).Get(ctx, &policyResult)
	if err != nil {
		logger.Error("Failed to ingest access policies", "error", err)
		return nil, err
	}
	result.PolicyCount = policyResult.PolicyCount

	// Phase 2: Ingest access levels and service perimeters in parallel (both depend on policies being in DB)
	levelFuture := workflow.ExecuteChildWorkflow(childCtx, accesslevel.GCPAccessContextManagerAccessLevelWorkflow,
		accesslevel.GCPAccessContextManagerAccessLevelWorkflowParams{})

	perimeterFuture := workflow.ExecuteChildWorkflow(childCtx, serviceperimeter.GCPAccessContextManagerServicePerimeterWorkflow,
		serviceperimeter.GCPAccessContextManagerServicePerimeterWorkflowParams{})

	var levelResult accesslevel.GCPAccessContextManagerAccessLevelWorkflowResult
	err = levelFuture.Get(ctx, &levelResult)
	if err != nil {
		logger.Error("Failed to ingest access levels", "error", err)
		return nil, err
	}
	result.LevelCount = levelResult.LevelCount

	var perimeterResult serviceperimeter.GCPAccessContextManagerServicePerimeterWorkflowResult
	err = perimeterFuture.Get(ctx, &perimeterResult)
	if err != nil {
		logger.Error("Failed to ingest service perimeters", "error", err)
		return nil, err
	}
	result.PerimeterCount = perimeterResult.PerimeterCount

	logger.Info("Completed GCPAccessContextManagerWorkflow",
		"policyCount", result.PolicyCount,
		"levelCount", result.LevelCount,
		"perimeterCount", result.PerimeterCount,
	)

	return result, nil
}
