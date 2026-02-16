package attestor

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPBinaryAuthorizationAttestorWorkflowParams contains parameters for the attestor workflow.
type GCPBinaryAuthorizationAttestorWorkflowParams struct {
	ProjectID string
}

// GCPBinaryAuthorizationAttestorWorkflowResult contains the result of the attestor workflow.
type GCPBinaryAuthorizationAttestorWorkflowResult struct {
	ProjectID      string
	AttestorCount  int
	DurationMillis int64
}

// GCPBinaryAuthorizationAttestorWorkflow ingests Binary Authorization attestors for a single project.
func GCPBinaryAuthorizationAttestorWorkflow(ctx workflow.Context, params GCPBinaryAuthorizationAttestorWorkflowParams) (*GCPBinaryAuthorizationAttestorWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPBinaryAuthorizationAttestorWorkflow", "projectID", params.ProjectID)

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

	var result IngestBinaryAuthorizationAttestorsResult
	err := workflow.ExecuteActivity(activityCtx, IngestBinaryAuthorizationAttestorsActivity, IngestBinaryAuthorizationAttestorsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest binary authorization attestors", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPBinaryAuthorizationAttestorWorkflow",
		"projectID", params.ProjectID,
		"attestorCount", result.AttestorCount,
	)

	return &GCPBinaryAuthorizationAttestorWorkflowResult{
		ProjectID:      result.ProjectID,
		AttestorCount:  result.AttestorCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
