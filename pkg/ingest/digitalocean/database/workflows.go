package database

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DODatabaseWorkflowResult contains the result of the Database workflow.
type DODatabaseWorkflowResult struct {
	ClusterCount      int
	FirewallRuleCount int
	UserCount         int
	ReplicaCount      int
	BackupCount       int
	ConfigCount       int
	PoolCount         int
	DurationMillis    int64
}

// DODatabaseWorkflow ingests DigitalOcean Database clusters and their child resources.
func DODatabaseWorkflow(ctx workflow.Context) (*DODatabaseWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DODatabaseWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	// Step 1: Ingest database clusters
	var clustersResult IngestDODatabasesResult
	err := workflow.ExecuteActivity(activityCtx, IngestDODatabasesActivity).Get(ctx, &clustersResult)
	if err != nil {
		logger.Error("Failed to ingest database clusters", "error", err)
		return nil, err
	}

	// Step 2: Ingest child resources using the cluster IDs from step 1
	var childrenResult IngestDODatabaseChildrenResult
	if len(clustersResult.ClusterIDs) > 0 {
		err = workflow.ExecuteActivity(activityCtx, IngestDODatabaseChildrenActivity, IngestDODatabaseChildrenInput{
			ClusterIDs: clustersResult.ClusterIDs,
			EngineMap:  clustersResult.EngineMap,
		}).Get(ctx, &childrenResult)
		if err != nil {
			logger.Error("Failed to ingest database children", "error", err)
			return nil, err
		}
	}

	logger.Info("Completed DODatabaseWorkflow",
		"clusterCount", clustersResult.ClusterCount,
		"firewallRuleCount", childrenResult.FirewallRuleCount,
		"userCount", childrenResult.UserCount,
		"replicaCount", childrenResult.ReplicaCount,
		"backupCount", childrenResult.BackupCount,
		"configCount", childrenResult.ConfigCount,
		"poolCount", childrenResult.PoolCount,
	)

	return &DODatabaseWorkflowResult{
		ClusterCount:      clustersResult.ClusterCount,
		FirewallRuleCount: childrenResult.FirewallRuleCount,
		UserCount:         childrenResult.UserCount,
		ReplicaCount:      childrenResult.ReplicaCount,
		BackupCount:       childrenResult.BackupCount,
		ConfigCount:       childrenResult.ConfigCount,
		PoolCount:         childrenResult.PoolCount,
		DurationMillis:    clustersResult.DurationMillis + childrenResult.DurationMillis,
	}, nil
}
