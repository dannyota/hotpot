package vault

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/vault/pki"
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

	// Process each vault instance
	for _, vaultName := range listResult.VaultNames {
		instanceResult := InstanceResult{VaultName: vaultName}

		var pkiResult pki.VaultPKIWorkflowResult
		err := workflow.ExecuteChildWorkflow(ctx, pki.VaultPKIWorkflow, pki.VaultPKIWorkflowParams{
			VaultName: vaultName,
		}).Get(ctx, &pkiResult)

		if err != nil {
			logger.Error("Failed to execute VaultPKIWorkflow", "vaultName", vaultName, "error", err)
			instanceResult.Error = err.Error()
		} else {
			instanceResult.CertificateCount = pkiResult.TotalCertificates
			result.TotalCertificates += pkiResult.TotalCertificates
		}

		result.InstanceResults = append(result.InstanceResults, instanceResult)
	}

	logger.Info("Completed VaultInventoryWorkflow",
		"totalCertificates", result.TotalCertificates,
		"instanceCount", len(listResult.VaultNames),
	)

	return result, nil
}
