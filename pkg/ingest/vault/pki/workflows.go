package pki

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/vault/pki/certificate"
)

// VaultPKIWorkflowParams contains parameters for the PKI workflow.
type VaultPKIWorkflowParams struct {
	VaultName string
}

// VaultPKIWorkflowResult contains the result of the PKI workflow.
type VaultPKIWorkflowResult struct {
	TotalCertificates int
	MountResults      []MountResult
}

// MountResult contains the ingestion result for a single PKI mount.
type MountResult struct {
	MountPath        string
	CertificateCount int
	Error            string
}

// VaultPKIWorkflow discovers PKI mounts and fans out certificate ingestion.
func VaultPKIWorkflow(ctx workflow.Context, params VaultPKIWorkflowParams) (*VaultPKIWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting VaultPKIWorkflow", "vaultName", params.VaultName)

	// Activity options for mount discovery
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	// Discover PKI mounts
	var discoverResult DiscoverMountsResult
	err := workflow.ExecuteActivity(activityCtx, DiscoverMountsActivity, DiscoverMountsParams{
		VaultName: params.VaultName,
	}).Get(ctx, &discoverResult)
	if err != nil {
		logger.Error("Failed to discover PKI mounts", "vaultName", params.VaultName, "error", err)
		return nil, err
	}

	// Child workflow options
	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 20 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, childOpts)

	result := &VaultPKIWorkflowResult{
		MountResults: make([]MountResult, 0, len(discoverResult.MountPaths)),
	}

	// Fan out certificate ingestion per mount
	for _, mountPath := range discoverResult.MountPaths {
		mountResult := MountResult{MountPath: mountPath}

		var certResult certificate.VaultPKICertificateWorkflowResult
		err := workflow.ExecuteChildWorkflow(ctx, certificate.VaultPKICertificateWorkflow, certificate.VaultPKICertificateWorkflowParams{
			VaultName: params.VaultName,
			MountPath: mountPath,
		}).Get(ctx, &certResult)

		if err != nil {
			logger.Error("Failed to execute VaultPKICertificateWorkflow",
				"vaultName", params.VaultName,
				"mountPath", mountPath,
				"error", err,
			)
			mountResult.Error = err.Error()
		} else {
			mountResult.CertificateCount = certResult.CertificateCount
			result.TotalCertificates += certResult.CertificateCount
		}

		result.MountResults = append(result.MountResults, mountResult)
	}

	logger.Info("Completed VaultPKIWorkflow",
		"vaultName", params.VaultName,
		"totalCertificates", result.TotalCertificates,
		"mountCount", len(discoverResult.MountPaths),
	)

	return result, nil
}
