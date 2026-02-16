package instance

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPFilestoreInstanceWorkflowParams contains parameters for the Filestore instance workflow.
type GCPFilestoreInstanceWorkflowParams struct {
	ProjectID string
}

// GCPFilestoreInstanceWorkflowResult contains the result of the Filestore instance workflow.
type GCPFilestoreInstanceWorkflowResult struct {
	ProjectID      string
	InstanceCount  int
	DurationMillis int64
}

// GCPFilestoreInstanceWorkflow ingests GCP Filestore instances for a single project.
func GCPFilestoreInstanceWorkflow(ctx workflow.Context, params GCPFilestoreInstanceWorkflowParams) (*GCPFilestoreInstanceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPFilestoreInstanceWorkflow", "projectID", params.ProjectID)

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

	var result IngestFilestoreInstancesResult
	err := workflow.ExecuteActivity(activityCtx, IngestFilestoreInstancesActivity, IngestFilestoreInstancesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Filestore instances", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPFilestoreInstanceWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
	)

	return &GCPFilestoreInstanceWorkflowResult{
		ProjectID:      result.ProjectID,
		InstanceCount:  result.InstanceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
