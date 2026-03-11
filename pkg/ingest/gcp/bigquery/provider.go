package bigquery

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "bigquery",
		Scope:     ingest.ScopeRegional,
		APIName:   "bigquery.googleapis.com",
		Register:  Register,
		Workflow:  GCPBigQueryWorkflow,
		NewParams: func(projectID, _, _ string) any {
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
