package kms

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "kms",
		Scope:     ingest.ScopeRegional,
		APIName:   "cloudkms.googleapis.com",
		Register:  Register,
		Workflow:  GCPKMSWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPKMSWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPKMSWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPKMSWorkflowResult)
			pr.KeyRingCount = r.KeyRingCount
			pr.CryptoKeyCount = r.CryptoKeyCount
			result.TotalKeyRings += r.KeyRingCount
			result.TotalCryptoKeys += r.CryptoKeyCount
		},
	})
}
