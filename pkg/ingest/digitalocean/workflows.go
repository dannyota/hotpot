package digitalocean

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest"
)

// DOInventoryWorkflowResult contains the result of DigitalOcean inventory collection.
type DOInventoryWorkflowResult struct {
	AccountCount              int
	DatabaseClusterCount      int
	DatabaseFirewallRuleCount int
	DatabaseUserCount         int
	DatabaseReplicaCount      int
	DatabaseBackupCount       int
	DatabaseConfigCount       int
	DatabasePoolCount         int
	DomainCount               int
	DomainRecordCount         int
	DropletCount              int
	FirewallCount             int
	KeyCount                  int
	KubernetesClusterCount    int
	KubernetesNodePoolCount   int
	LoadBalancerCount         int
	ProjectCount              int
	ResourceCount             int
	VolumeCount               int
	VpcCount                  int
}

// aggregateFunc is the function signature for merging a service result into the provider result.
type aggregateFunc = func(*DOInventoryWorkflowResult, any)

// DOInventoryWorkflow orchestrates DigitalOcean inventory collection.
func DOInventoryWorkflow(ctx workflow.Context) (*DOInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DOInventoryWorkflow")

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 60 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, childOpts)

	result := &DOInventoryWorkflowResult{}

	for _, svc := range ingest.Services("digitalocean") {
		res := svc.NewResult()
		err := workflow.ExecuteChildWorkflow(ctx, svc.Workflow).Get(ctx, res)
		if err != nil {
			logger.Error("Failed ingestion", "service", svc.Name, "error", err)
		} else {
			svc.Aggregate.(aggregateFunc)(result, res)
		}
	}

	logger.Info("Completed DOInventoryWorkflow",
		"accounts", result.AccountCount,
		"databaseClusters", result.DatabaseClusterCount,
		"databaseFirewallRules", result.DatabaseFirewallRuleCount,
		"databaseUsers", result.DatabaseUserCount,
		"databaseReplicas", result.DatabaseReplicaCount,
		"databaseBackups", result.DatabaseBackupCount,
		"databaseConfigs", result.DatabaseConfigCount,
		"databasePools", result.DatabasePoolCount,
		"domains", result.DomainCount,
		"domainRecords", result.DomainRecordCount,
		"droplets", result.DropletCount,
		"firewalls", result.FirewallCount,
		"keys", result.KeyCount,
		"kubernetesClusters", result.KubernetesClusterCount,
		"kubernetesNodePools", result.KubernetesNodePoolCount,
		"loadBalancers", result.LoadBalancerCount,
		"projects", result.ProjectCount,
		"projectResources", result.ResourceCount,
		"volumes", result.VolumeCount,
		"vpcs", result.VpcCount,
	)

	return result, nil
}
