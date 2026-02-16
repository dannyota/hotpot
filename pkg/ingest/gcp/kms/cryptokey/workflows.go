package cryptokey

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPKMSCryptoKeyWorkflowParams contains parameters for the crypto key workflow.
type GCPKMSCryptoKeyWorkflowParams struct {
	ProjectID string
}

// GCPKMSCryptoKeyWorkflowResult contains the result of the crypto key workflow.
type GCPKMSCryptoKeyWorkflowResult struct {
	ProjectID      string
	CryptoKeyCount int
	DurationMillis int64
}

// GCPKMSCryptoKeyWorkflow ingests GCP KMS crypto keys for a single project.
func GCPKMSCryptoKeyWorkflow(ctx workflow.Context, params GCPKMSCryptoKeyWorkflowParams) (*GCPKMSCryptoKeyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPKMSCryptoKeyWorkflow", "projectID", params.ProjectID)

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

	var result IngestKMSCryptoKeysResult
	err := workflow.ExecuteActivity(activityCtx, IngestKMSCryptoKeysActivity, IngestKMSCryptoKeysParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest crypto keys", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPKMSCryptoKeyWorkflow",
		"projectID", params.ProjectID,
		"cryptoKeyCount", result.CryptoKeyCount,
	)

	return &GCPKMSCryptoKeyWorkflowResult{
		ProjectID:      result.ProjectID,
		CryptoKeyCount: result.CryptoKeyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
