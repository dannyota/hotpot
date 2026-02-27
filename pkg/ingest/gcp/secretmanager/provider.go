package secretmanager

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "secretmanager",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPSecretManagerWorkflow,
		NewParams: func(projectID, _ string) any {
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
