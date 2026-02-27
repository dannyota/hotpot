package containeranalysis

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "containeranalysis",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPContainerAnalysisWorkflow,
		NewParams: func(projectID, _ string) any {
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
