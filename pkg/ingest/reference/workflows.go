package reference

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest"
)

// ReferenceInventoryWorkflowResult contains the result of reference data collection.
type ReferenceInventoryWorkflowResult struct {
	CPECount           int
	UbuntuPackageCount int
	RPMPackageCount    int
	EOLProductCount    int
	EOLCycleCount      int
	XeolProductCount   int
	XeolCycleCount     int
	XeolPurlCount      int
	XeolVulnCount      int
}

// aggregateFunc is the function signature for merging a service result into the provider result.
type aggregateFunc = func(*ReferenceInventoryWorkflowResult, any)

// ReferenceInventoryWorkflow orchestrates reference data collection.
func ReferenceInventoryWorkflow(ctx workflow.Context) (*ReferenceInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting ReferenceInventoryWorkflow")

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 120 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, childOpts)

	result := &ReferenceInventoryWorkflowResult{}

	for _, svc := range ingest.Services("reference") {
		res := svc.NewResult()
		err := workflow.ExecuteChildWorkflow(ctx, svc.Workflow).Get(ctx, res)
		if err != nil {
			logger.Error("Failed ingestion", "service", svc.Name, "error", err)
		} else {
			svc.Aggregate.(aggregateFunc)(result, res)
		}
	}

	logger.Info("Completed ReferenceInventoryWorkflow",
		"cpeCount", result.CPECount,
		"ubuntuPackages", result.UbuntuPackageCount,
		"rpmPackages", result.RPMPackageCount,
		"eolProducts", result.EOLProductCount,
		"eolCycles", result.EOLCycleCount,
		"xeolProducts", result.XeolProductCount,
		"xeolCycles", result.XeolCycleCount,
		"xeolPurls", result.XeolPurlCount,
		"xeolVulns", result.XeolVulnCount,
	)

	return result, nil
}
