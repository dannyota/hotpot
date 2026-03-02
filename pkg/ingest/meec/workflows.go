package meec

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest"
)

// MEECInventoryWorkflowResult contains the result of MEEC inventory collection.
type MEECInventoryWorkflowResult struct {
	ComputerCount          int
	SoftwareCount          int
	InstalledSoftwareCount int
}

// aggregateFunc is the function signature for merging a service result into the provider result.
type aggregateFunc = func(*MEECInventoryWorkflowResult, any)

// MEECInventoryWorkflow orchestrates MEEC inventory collection.
func MEECInventoryWorkflow(ctx workflow.Context) (*MEECInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting MEECInventoryWorkflow")

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

	result := &MEECInventoryWorkflowResult{}

	for _, svc := range ingest.Services("meec") {
		res := svc.NewResult()
		err := workflow.ExecuteChildWorkflow(ctx, svc.Workflow).Get(ctx, res)
		if err != nil {
			logger.Error("Failed ingestion", "service", svc.Name, "error", err)
		} else {
			svc.Aggregate.(aggregateFunc)(result, res)
		}
	}

	logger.Info("Completed MEECInventoryWorkflow",
		"computers", result.ComputerCount,
		"software", result.SoftwareCount,
		"installedSoftware", result.InstalledSoftwareCount,
	)

	return result, nil
}
