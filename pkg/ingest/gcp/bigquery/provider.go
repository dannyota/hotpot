package bigquery

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "bigquery",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPBigQueryWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPBigQueryWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPBigQueryWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPBigQueryWorkflowResult)
			pr.DatasetCount = r.DatasetCount
			pr.TableCount = r.TableCount
			result.TotalDatasets += r.DatasetCount
			result.TotalTables += r.TableCount
		},
	})
}
