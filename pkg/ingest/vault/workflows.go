package vault

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest"
)

// VaultInventoryWorkflowResult contains the result of the Vault inventory workflow.
type VaultInventoryWorkflowResult struct {
	TotalCertificates int
	InstanceResults   []InstanceResult
}

// InstanceResult contains the ingestion result for a single Vault instance.
type InstanceResult struct {
	VaultName        string
	CertificateCount int
	Error            string
}

// aggregateFunc is the function signature for merging a service result into the provider-level results.
type aggregateFunc = func(*VaultInventoryWorkflowResult, *InstanceResult, any)

// VaultInventoryWorkflow ingests PKI certificates from all configured Vault instances.
func VaultInventoryWorkflow(ctx workflow.Context) (*VaultInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting VaultInventoryWorkflow")

	// Activity options for listing instances
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	// List vault instances from config
	var listResult ListVaultInstancesResult
	err := workflow.ExecuteActivity(activityCtx, ListVaultInstancesActivity).Get(ctx, &listResult)
	if err != nil {
		logger.Error("Failed to list vault instances", "error", err)
		return nil, err
	}

	// Child workflow options
	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, childOpts)

	result := &VaultInventoryWorkflowResult{
		InstanceResults: make([]InstanceResult, 0, len(listResult.VaultNames)),
	}

	services := ingest.Services("vault")

	// Process each vault instance
	for _, vaultName := range listResult.VaultNames {
		instanceResult := InstanceResult{VaultName: vaultName}

		for _, svc := range services {
			res := svc.NewResult()
			err := workflow.ExecuteChildWorkflow(ctx, svc.Workflow,
				svc.NewParams(vaultName, "")).Get(ctx, res)
			if err != nil {
				logger.Error("Failed ingestion", "service", svc.Name, "vaultName", vaultName, "error", err)
				appendError(&instanceResult, err)
			} else {
				svc.Aggregate.(aggregateFunc)(result, &instanceResult, res)
			}
		}

		result.InstanceResults = append(result.InstanceResults, instanceResult)
	}

	logger.Info("Completed VaultInventoryWorkflow",
		"totalCertificates", result.TotalCertificates,
		"instanceCount", len(listResult.VaultNames),
	)

	return result, nil
}

func appendError(ir *InstanceResult, err error) {
	if ir.Error == "" {
		ir.Error = err.Error()
	} else {
		ir.Error += "; " + err.Error()
	}
}
