package secgroup

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeNetworkSecgroupWorkflowParams contains parameters for the security group workflow.
type GreenNodeNetworkSecgroupWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeNetworkSecgroupWorkflowResult contains the result of the security group workflow.
type GreenNodeNetworkSecgroupWorkflowResult struct {
	SecgroupCount  int
	DurationMillis int64
}

// GreenNodeNetworkSecgroupWorkflow ingests GreenNode security groups.
func GreenNodeNetworkSecgroupWorkflow(ctx workflow.Context, params GreenNodeNetworkSecgroupWorkflowParams) (*GreenNodeNetworkSecgroupWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeNetworkSecgroupWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestNetworkSecgroupsResult
	err := workflow.ExecuteActivity(activityCtx, IngestNetworkSecgroupsActivity, IngestNetworkSecgroupsParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest secgroups", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeNetworkSecgroupWorkflow",
		"secgroupCount", result.SecgroupCount,
	)

	return &GreenNodeNetworkSecgroupWorkflowResult{
		SecgroupCount:  result.SecgroupCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
