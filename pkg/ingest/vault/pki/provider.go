package pki

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/vault"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "vault",
		Name:      "pki",
		Register:  Register,
		Workflow:  VaultPKIWorkflow,
		NewParams: func(vaultName, _, _ string) any {
			return VaultPKIWorkflowParams{VaultName: vaultName}
		},
		NewResult: func() any { return &VaultPKIWorkflowResult{} },
		Aggregate: func(result *vault.VaultInventoryWorkflowResult, ir *vault.InstanceResult, child any) {
			r := child.(*VaultPKIWorkflowResult)
			ir.CertificateCount = r.TotalCertificates
			result.TotalCertificates += r.TotalCertificates
		},
	})
}
