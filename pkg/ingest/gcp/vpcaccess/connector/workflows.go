package connector

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPVpcAccessConnectorWorkflowParams contains parameters for the connector workflow.
type GCPVpcAccessConnectorWorkflowParams struct {
	ProjectID string
}

// GCPVpcAccessConnectorWorkflowResult contains the result of the connector workflow.
type GCPVpcAccessConnectorWorkflowResult struct {
	ProjectID      string
	ConnectorCount int
	DurationMillis int64
}

// GCPVpcAccessConnectorWorkflow ingests GCP VPC Access connectors for a single project.
func GCPVpcAccessConnectorWorkflow(ctx workflow.Context, params GCPVpcAccessConnectorWorkflowParams) (*GCPVpcAccessConnectorWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPVpcAccessConnectorWorkflow", "projectID", params.ProjectID)

	// Activity options
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

	// Execute ingest activity
	var result IngestVpcAccessConnectorsResult
	err := workflow.ExecuteActivity(activityCtx, IngestVpcAccessConnectorsActivity, IngestVpcAccessConnectorsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest VPC Access connectors", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPVpcAccessConnectorWorkflow",
		"projectID", params.ProjectID,
		"connectorCount", result.ConnectorCount,
	)

	return &GCPVpcAccessConnectorWorkflowResult{
		ProjectID:      result.ProjectID,
		ConnectorCount: result.ConnectorCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
