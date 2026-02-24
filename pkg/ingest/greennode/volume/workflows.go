package volume

import (
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/greennode/volume/blockvolume"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/volume/volumetype"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/volume/volumetypezone"
)

// GreenNodeVolumeWorkflowParams contains parameters for the volume workflow.
type GreenNodeVolumeWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeVolumeWorkflowResult contains the result of the volume workflow.
type GreenNodeVolumeWorkflowResult struct {
	BlockVolumeCount    int
	VolumeTypeCount     int
	VolumeTypeZoneCount int
}

// GreenNodeVolumeWorkflow orchestrates GreenNode volume ingestion.
func GreenNodeVolumeWorkflow(ctx workflow.Context, params GreenNodeVolumeWorkflowParams) (*GreenNodeVolumeWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeVolumeWorkflow", "projectID", params.ProjectID, "region", params.Region)

	result := &GreenNodeVolumeWorkflowResult{}

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Block Volumes
	var blockVolumeResult blockvolume.GreenNodeVolumeBlockVolumeWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, blockvolume.GreenNodeVolumeBlockVolumeWorkflow, blockvolume.GreenNodeVolumeBlockVolumeWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &blockVolumeResult)
	if err != nil {
		logger.Error("Failed to ingest block volumes", "error", err)
	} else {
		result.BlockVolumeCount = blockVolumeResult.BlockVolumeCount
	}

	// Volume Types
	var volumeTypeResult volumetype.GreenNodeVolumeVolumeTypeWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, volumetype.GreenNodeVolumeVolumeTypeWorkflow, volumetype.GreenNodeVolumeVolumeTypeWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &volumeTypeResult)
	if err != nil {
		logger.Error("Failed to ingest volume types", "error", err)
	} else {
		result.VolumeTypeCount = volumeTypeResult.VolumeTypeCount
	}

	// Volume Type Zones
	var volumeTypeZoneResult volumetypezone.GreenNodeVolumeVolumeTypeZoneWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, volumetypezone.GreenNodeVolumeVolumeTypeZoneWorkflow, volumetypezone.GreenNodeVolumeVolumeTypeZoneWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &volumeTypeZoneResult)
	if err != nil {
		logger.Error("Failed to ingest volume type zones", "error", err)
	} else {
		result.VolumeTypeZoneCount = volumeTypeZoneResult.VolumeTypeZoneCount
	}

	logger.Info("Completed GreenNodeVolumeWorkflow",
		"blockVolumeCount", result.BlockVolumeCount,
		"volumeTypeCount", result.VolumeTypeCount,
		"volumeTypeZoneCount", result.VolumeTypeZoneCount,
	)

	return result, nil
}
