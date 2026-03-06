package compute

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "compute",
		Scope:     ingest.ScopeRegional,
		APIName:   "compute.googleapis.com",
		Register:  Register,
		Workflow:  GCPComputeWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPComputeWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPComputeWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPComputeWorkflowResult)
			pr.InstanceCount = r.InstanceCount
			pr.InterconnectCount = r.InterconnectCount
			pr.PacketMirroringCount = r.PacketMirroringCount
			pr.ProjectMetadataCount = r.ProjectMetadataCount
			result.TotalInstances += r.InstanceCount
			result.TotalInterconnects += r.InterconnectCount
			result.TotalPacketMirrorings += r.PacketMirroringCount
			result.TotalProjectMetadata += r.ProjectMetadataCount
		},
	})
}
