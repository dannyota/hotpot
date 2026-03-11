package containeranalysis

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "containeranalysis",
		Scope:     ingest.ScopeRegional,
		APIName:   "containeranalysis.googleapis.com",
		Register:  Register,
		Workflow:  GCPContainerAnalysisWorkflow,
		NewParams: func(projectID, _, _ string) any {
			return GCPContainerAnalysisWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPContainerAnalysisWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPContainerAnalysisWorkflowResult)
			pr.NoteCount = r.NoteCount
			pr.OccurrenceCount = r.OccurrenceCount
			result.TotalNotes += r.NoteCount
			result.TotalOccurrences += r.OccurrenceCount
		},
	})
}
