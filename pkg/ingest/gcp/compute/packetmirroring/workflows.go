package packetmirroring

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputePacketMirroringWorkflowParams contains parameters for the packet mirroring workflow.
type GCPComputePacketMirroringWorkflowParams struct {
	ProjectID string
}

// GCPComputePacketMirroringWorkflowResult contains the result of the packet mirroring workflow.
type GCPComputePacketMirroringWorkflowResult struct {
	ProjectID            string
	PacketMirroringCount int
	DurationMillis       int64
}

// GCPComputePacketMirroringWorkflow ingests GCP Compute packet mirrorings for a single project.
func GCPComputePacketMirroringWorkflow(ctx workflow.Context, params GCPComputePacketMirroringWorkflowParams) (*GCPComputePacketMirroringWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputePacketMirroringWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputePacketMirroringsResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputePacketMirroringsActivity, IngestComputePacketMirroringsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest packet mirrorings", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputePacketMirroringWorkflow",
		"projectID", params.ProjectID,
		"packetMirroringCount", result.PacketMirroringCount,
	)

	return &GCPComputePacketMirroringWorkflowResult{
		ProjectID:            result.ProjectID,
		PacketMirroringCount: result.PacketMirroringCount,
		DurationMillis:       result.DurationMillis,
	}, nil
}
