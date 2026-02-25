package vpc

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeNetworkVPCWorkflowParams contains parameters for the VPC workflow.
type GreenNodeNetworkVPCWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeNetworkVPCWorkflowResult contains the result of the VPC workflow.
type GreenNodeNetworkVPCWorkflowResult struct {
	VPCCount       int
	DurationMillis int64
}

// GreenNodeNetworkVPCWorkflow ingests GreenNode VPCs.
func GreenNodeNetworkVPCWorkflow(ctx workflow.Context, params GreenNodeNetworkVPCWorkflowParams) (*GreenNodeNetworkVPCWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeNetworkVPCWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestNetworkVPCsResult
	err := workflow.ExecuteActivity(activityCtx, IngestNetworkVPCsActivity, IngestNetworkVPCsParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest vpcs", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeNetworkVPCWorkflow", "vpcCount", result.VPCCount)

	return &GreenNodeNetworkVPCWorkflowResult{
		VPCCount:       result.VPCCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
