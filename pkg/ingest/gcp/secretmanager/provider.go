package secretmanager

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "secretmanager",
		Scope:     ingest.ScopeRegional,
		APIName:   "secretmanager.googleapis.com",
		Register:  Register,
		Workflow:  GCPSecretManagerWorkflow,
		NewParams: func(projectID, _, _ string) any {
			return GCPSecretManagerWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPSecretManagerWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPSecretManagerWorkflowResult)
			pr.SecretCount = r.SecretCount
			result.TotalSecrets += r.SecretCount
		},
	})
}
