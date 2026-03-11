package certificate

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// VaultPKICertificateWorkflowParams contains parameters for the certificate workflow.
type VaultPKICertificateWorkflowParams struct {
	VaultName string
	MountPath string
}

// VaultPKICertificateWorkflowResult contains the result of the certificate workflow.
type VaultPKICertificateWorkflowResult struct {
	CertificateCount int
	DurationMillis   int64
}

// VaultPKICertificateWorkflow ingests certificates from a single Vault PKI mount.
func VaultPKICertificateWorkflow(ctx workflow.Context, params VaultPKICertificateWorkflowParams) (*VaultPKICertificateWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting VaultPKICertificateWorkflow",
		"vaultName", params.VaultName,
		"mountPath", params.MountPath,
	)

	// Activity options
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	// Execute ingest activity
	var result IngestCertificatesResult
	err := workflow.ExecuteActivity(activityCtx, IngestCertificatesActivity, IngestCertificatesParams{
		VaultName: params.VaultName,
		MountPath: params.MountPath,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest certificates", "error", err)
		return nil, err
	}

	logger.Info("Completed VaultPKICertificateWorkflow",
		"vaultName", params.VaultName,
		"mountPath", params.MountPath,
		"certificateCount", result.CertificateCount,
	)

	return &VaultPKICertificateWorkflowResult{
		CertificateCount: result.CertificateCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
