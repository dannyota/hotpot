package kubernetes

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DOKubernetesWorkflowResult contains the result of the Kubernetes workflow.
type DOKubernetesWorkflowResult struct {
	ClusterCount   int
	NodePoolCount  int
	DurationMillis int64
}

// DOKubernetesWorkflow ingests DigitalOcean Kubernetes clusters and their node pools.
func DOKubernetesWorkflow(ctx workflow.Context) (*DOKubernetesWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DOKubernetesWorkflow")

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

	// Step 1: Ingest Kubernetes clusters
	var clustersResult IngestDOKubernetesClustersResult
	err := workflow.ExecuteActivity(activityCtx, IngestDOKubernetesClustersActivity).Get(ctx, &clustersResult)
	if err != nil {
		logger.Error("Failed to ingest Kubernetes clusters", "error", err)
		return nil, err
	}

	// Step 2: Ingest node pools using the cluster IDs from step 1
	var nodePoolsResult IngestDOKubernetesNodePoolsResult
	if len(clustersResult.ClusterIDs) > 0 {
		err = workflow.ExecuteActivity(activityCtx, IngestDOKubernetesNodePoolsActivity, IngestDOKubernetesNodePoolsInput{
			ClusterIDs: clustersResult.ClusterIDs,
		}).Get(ctx, &nodePoolsResult)
		if err != nil {
			logger.Error("Failed to ingest Kubernetes node pools", "error", err)
			return nil, err
		}
	}

	logger.Info("Completed DOKubernetesWorkflow",
		"clusterCount", clustersResult.ClusterCount,
		"nodePoolCount", nodePoolsResult.NodePoolCount,
	)

	return &DOKubernetesWorkflowResult{
		ClusterCount:   clustersResult.ClusterCount,
		NodePoolCount:  nodePoolsResult.NodePoolCount,
		DurationMillis: clustersResult.DurationMillis + nodePoolsResult.DurationMillis,
	}, nil
}
