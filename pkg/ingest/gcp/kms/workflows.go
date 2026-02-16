package kms

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/kms/cryptokey"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/kms/keyring"
)

// GCPKMSWorkflowParams contains parameters for the KMS workflow.
type GCPKMSWorkflowParams struct {
	ProjectID string
}

// GCPKMSWorkflowResult contains the result of the KMS workflow.
type GCPKMSWorkflowResult struct {
	ProjectID      string
	KeyRingCount   int
	CryptoKeyCount int
}

// GCPKMSWorkflow ingests all GCP KMS resources for a single project.
// KeyRings are ingested first, then CryptoKeys (which depend on key ring data).
func GCPKMSWorkflow(ctx workflow.Context, params GCPKMSWorkflowParams) (*GCPKMSWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPKMSWorkflow", "projectID", params.ProjectID)

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPKMSWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// KeyRings first (CryptoKeys depend on them)
	var keyRingResult keyring.GCPKMSKeyRingWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, keyring.GCPKMSKeyRingWorkflow,
		keyring.GCPKMSKeyRingWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &keyRingResult)
	if err != nil {
		logger.Error("Failed to ingest key rings", "error", err)
		return nil, err
	}
	result.KeyRingCount = keyRingResult.KeyRingCount

	// CryptoKeys (queries key rings from database)
	var cryptoKeyResult cryptokey.GCPKMSCryptoKeyWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, cryptokey.GCPKMSCryptoKeyWorkflow,
		cryptokey.GCPKMSCryptoKeyWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &cryptoKeyResult)
	if err != nil {
		logger.Error("Failed to ingest crypto keys", "error", err)
		return nil, err
	}
	result.CryptoKeyCount = cryptoKeyResult.CryptoKeyCount

	logger.Info("Completed GCPKMSWorkflow",
		"projectID", params.ProjectID,
		"keyRingCount", result.KeyRingCount,
		"cryptoKeyCount", result.CryptoKeyCount,
	)

	return result, nil
}
