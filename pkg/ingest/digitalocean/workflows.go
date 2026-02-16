package digitalocean

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/account"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/database"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/domain"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/droplet"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/firewall"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/key"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/loadbalancer"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/project"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/volume"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean/vpc"
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
	LoadBalancerCount         int
	ProjectCount              int
	ResourceCount             int
	VolumeCount               int
	VpcCount                  int
}

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

	// Launch all child workflows concurrently
	accountFuture := workflow.ExecuteChildWorkflow(ctx, account.DOAccountWorkflow)
	databaseFuture := workflow.ExecuteChildWorkflow(ctx, database.DODatabaseWorkflow)
	domainFuture := workflow.ExecuteChildWorkflow(ctx, domain.DODomainWorkflow)
	dropletFuture := workflow.ExecuteChildWorkflow(ctx, droplet.DODropletWorkflow)
	firewallFuture := workflow.ExecuteChildWorkflow(ctx, firewall.DOFirewallWorkflow)
	keyFuture := workflow.ExecuteChildWorkflow(ctx, key.DOKeyWorkflow)
	lbFuture := workflow.ExecuteChildWorkflow(ctx, loadbalancer.DOLoadBalancerWorkflow)
	projectFuture := workflow.ExecuteChildWorkflow(ctx, project.DOProjectWorkflow)
	volumeFuture := workflow.ExecuteChildWorkflow(ctx, volume.DOVolumeWorkflow)
	vpcFuture := workflow.ExecuteChildWorkflow(ctx, vpc.DOVpcWorkflow)

	// Collect results
	var accountResult account.DOAccountWorkflowResult
	if err := accountFuture.Get(ctx, &accountResult); err != nil {
		logger.Error("Failed to execute DOAccountWorkflow", "error", err)
	} else {
		result.AccountCount = accountResult.AccountCount
	}

	var databaseResult database.DODatabaseWorkflowResult
	if err := databaseFuture.Get(ctx, &databaseResult); err != nil {
		logger.Error("Failed to execute DODatabaseWorkflow", "error", err)
	} else {
		result.DatabaseClusterCount = databaseResult.ClusterCount
		result.DatabaseFirewallRuleCount = databaseResult.FirewallRuleCount
		result.DatabaseUserCount = databaseResult.UserCount
		result.DatabaseReplicaCount = databaseResult.ReplicaCount
		result.DatabaseBackupCount = databaseResult.BackupCount
		result.DatabaseConfigCount = databaseResult.ConfigCount
		result.DatabasePoolCount = databaseResult.PoolCount
	}

	var domainResult domain.DODomainWorkflowResult
	if err := domainFuture.Get(ctx, &domainResult); err != nil {
		logger.Error("Failed to execute DODomainWorkflow", "error", err)
	} else {
		result.DomainCount = domainResult.DomainCount
		result.DomainRecordCount = domainResult.RecordCount
	}

	var dropletResult droplet.DODropletWorkflowResult
	if err := dropletFuture.Get(ctx, &dropletResult); err != nil {
		logger.Error("Failed to execute DODropletWorkflow", "error", err)
	} else {
		result.DropletCount = dropletResult.DropletCount
	}

	var firewallResult firewall.DOFirewallWorkflowResult
	if err := firewallFuture.Get(ctx, &firewallResult); err != nil {
		logger.Error("Failed to execute DOFirewallWorkflow", "error", err)
	} else {
		result.FirewallCount = firewallResult.FirewallCount
	}

	var keyResult key.DOKeyWorkflowResult
	if err := keyFuture.Get(ctx, &keyResult); err != nil {
		logger.Error("Failed to execute DOKeyWorkflow", "error", err)
	} else {
		result.KeyCount = keyResult.KeyCount
	}

	var lbResult loadbalancer.DOLoadBalancerWorkflowResult
	if err := lbFuture.Get(ctx, &lbResult); err != nil {
		logger.Error("Failed to execute DOLoadBalancerWorkflow", "error", err)
	} else {
		result.LoadBalancerCount = lbResult.LoadBalancerCount
	}

	var projectResult project.DOProjectWorkflowResult
	if err := projectFuture.Get(ctx, &projectResult); err != nil {
		logger.Error("Failed to execute DOProjectWorkflow", "error", err)
	} else {
		result.ProjectCount = projectResult.ProjectCount
		result.ResourceCount = projectResult.ResourceCount
	}

	var volumeResult volume.DOVolumeWorkflowResult
	if err := volumeFuture.Get(ctx, &volumeResult); err != nil {
		logger.Error("Failed to execute DOVolumeWorkflow", "error", err)
	} else {
		result.VolumeCount = volumeResult.VolumeCount
	}

	var vpcResult vpc.DOVpcWorkflowResult
	if err := vpcFuture.Get(ctx, &vpcResult); err != nil {
		logger.Error("Failed to execute DOVpcWorkflow", "error", err)
	} else {
		result.VpcCount = vpcResult.VpcCount
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
		"loadBalancers", result.LoadBalancerCount,
		"projects", result.ProjectCount,
		"projectResources", result.ResourceCount,
		"volumes", result.VolumeCount,
		"vpcs", result.VpcCount,
	)

	return result, nil
}
