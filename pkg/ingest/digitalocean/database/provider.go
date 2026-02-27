package database

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "digitalocean",
		Name:      "database",
		Register:  Register,
		Workflow:  DODatabaseWorkflow,
		NewResult: func() any { return &DODatabaseWorkflowResult{} },
		Aggregate: func(result *digitalocean.DOInventoryWorkflowResult, child any) {
			r := child.(*DODatabaseWorkflowResult)
			result.DatabaseClusterCount = r.ClusterCount
			result.DatabaseFirewallRuleCount = r.FirewallRuleCount
			result.DatabaseUserCount = r.UserCount
			result.DatabaseReplicaCount = r.ReplicaCount
			result.DatabaseBackupCount = r.BackupCount
			result.DatabaseConfigCount = r.ConfigCount
			result.DatabasePoolCount = r.PoolCount
		},
	})
}
