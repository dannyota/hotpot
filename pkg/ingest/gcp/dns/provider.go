package dns

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "dns",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPDNSWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPDNSWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPDNSWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPDNSWorkflowResult)
			pr.ManagedZoneCount = r.ManagedZoneCount
			pr.DNSPolicyCount = r.PolicyCount
			result.TotalManagedZones += r.ManagedZoneCount
			result.TotalDNSPolicies += r.PolicyCount
		},
	})
}
