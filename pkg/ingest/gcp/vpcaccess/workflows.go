package vpcaccess

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/vpcaccess/connector"
)

// GCPVpcAccessWorkflowParams contains parameters for the VPC Access workflow.
type GCPVpcAccessWorkflowParams struct {
	ProjectID string
}

// GCPVpcAccessWorkflowResult contains the result of the VPC Access workflow.
type GCPVpcAccessWorkflowResult struct {
	ProjectID      string
	ConnectorCount int
}

// GCPVpcAccessWorkflow ingests all VPC Access resources for a single project.
func GCPVpcAccessWorkflow(ctx workflow.Context, params GCPVpcAccessWorkflowParams) (*GCPVpcAccessWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPVpcAccessWorkflow", "projectID", params.ProjectID)

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

	result := &GCPVpcAccessWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Execute connector workflow
	var connectorResult connector.GCPVpcAccessConnectorWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, connector.GCPVpcAccessConnectorWorkflow,
		connector.GCPVpcAccessConnectorWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &connectorResult)
	if err != nil {
		logger.Error("Failed to ingest VPC Access connectors", "error", err)
		return nil, err
	}
	result.ConnectorCount = connectorResult.ConnectorCount

	logger.Info("Completed GCPVpcAccessWorkflow",
		"projectID", params.ProjectID,
		"connectorCount", result.ConnectorCount,
	)

	return result, nil
}
