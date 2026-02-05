package gcp

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"hotpot/pkg/ingest/gcp/compute"
)

// GCPInventoryWorkflowParams contains parameters for the GCP inventory workflow.
type GCPInventoryWorkflowParams struct {
	ProjectIDs []string
}

// GCPInventoryWorkflowResult contains the result of the GCP inventory workflow.
type GCPInventoryWorkflowResult struct {
	ProjectResults []ProjectResult
	TotalInstances int
	// TotalDisks     int  // future
	// TotalNetworks  int  // future
}

// ProjectResult contains the ingestion result for a single project.
type ProjectResult struct {
	ProjectID     string
	InstanceCount int
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

		// Execute GKEWorkflow for this project (future)
		// var gkeResult gke.GKEWorkflowResult
		// err = workflow.ExecuteChildWorkflow(ctx, gke.GKEWorkflow, gke.GKEWorkflowParams{...}).Get(ctx, &gkeResult)

		// Execute IAMWorkflow for this project (future)
		// var iamResult iam.IAMWorkflowResult
		// err = workflow.ExecuteChildWorkflow(ctx, iam.IAMWorkflow, iam.IAMWorkflowParams{...}).Get(ctx, &iamResult)

		result.ProjectResults = append(result.ProjectResults, projectResult)
	}

	logger.Info("Completed GCPInventoryWorkflow",
		"projectCount", len(params.ProjectIDs),
		"totalInstances", result.TotalInstances,
	)

	return result, nil
}
