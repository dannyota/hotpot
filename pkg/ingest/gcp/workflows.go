package gcp

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"hotpot/pkg/ingest/gcp/compute"
	"hotpot/pkg/ingest/gcp/container"
)

// GCPInventoryWorkflowParams contains parameters for the GCP inventory workflow.
type GCPInventoryWorkflowParams struct {
	ProjectIDs []string
}

// GCPInventoryWorkflowResult contains the result of the GCP inventory workflow.
type GCPInventoryWorkflowResult struct {
	ProjectResults []ProjectResult
	TotalInstances int
	TotalClusters  int
}

// ProjectResult contains the ingestion result for a single project.
type ProjectResult struct {
	ProjectID     string
	InstanceCount int
	ClusterCount  int
	Error         string
}

// GCPInventoryWorkflow ingests all GCP resources across multiple projects.
// It orchestrates compute, GKE, IAM, and other GCP resource ingestion.
func GCPInventoryWorkflow(ctx workflow.Context, params GCPInventoryWorkflowParams) (*GCPInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPInventoryWorkflow", "projectCount", len(params.ProjectIDs))

	// Child workflow options
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

	result := &GCPInventoryWorkflowResult{
		ProjectResults: make([]ProjectResult, 0, len(params.ProjectIDs)),
	}

	// Process each project
	for _, projectID := range params.ProjectIDs {
		projectResult := ProjectResult{ProjectID: projectID}

		// Execute GCPComputeWorkflow for this project
		var computeResult compute.GCPComputeWorkflowResult
		err := workflow.ExecuteChildWorkflow(ctx, compute.GCPComputeWorkflow, compute.GCPComputeWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &computeResult)

		if err != nil {
			logger.Error("Failed to execute GCPComputeWorkflow", "projectID", projectID, "error", err)
			projectResult.Error = err.Error()
		} else {
			projectResult.InstanceCount = computeResult.InstanceCount
			result.TotalInstances += computeResult.InstanceCount
		}

		// Execute GCPContainerWorkflow for this project
		var containerResult container.GCPContainerWorkflowResult
		err = workflow.ExecuteChildWorkflow(ctx, container.GCPContainerWorkflow, container.GCPContainerWorkflowParams{
			ProjectID: projectID,
		}).Get(ctx, &containerResult)

		if err != nil {
			logger.Error("Failed to execute GCPContainerWorkflow", "projectID", projectID, "error", err)
			if projectResult.Error == "" {
				projectResult.Error = err.Error()
			} else {
				projectResult.Error += "; " + err.Error()
			}
		} else {
			projectResult.ClusterCount = containerResult.ClusterCount
			result.TotalClusters += containerResult.ClusterCount
		}

		result.ProjectResults = append(result.ProjectResults, projectResult)
	}

	logger.Info("Completed GCPInventoryWorkflow",
		"projectCount", len(params.ProjectIDs),
		"totalInstances", result.TotalInstances,
		"totalClusters", result.TotalClusters,
	)

	return result, nil
}
