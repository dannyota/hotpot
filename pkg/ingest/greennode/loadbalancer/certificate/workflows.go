package certificate

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeLoadBalancerCertificateWorkflowParams contains parameters for the certificate workflow.
type GreenNodeLoadBalancerCertificateWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeLoadBalancerCertificateWorkflowResult contains the result of the certificate workflow.
type GreenNodeLoadBalancerCertificateWorkflowResult struct {
	CertificateCount int
	DurationMillis   int64
}

// GreenNodeLoadBalancerCertificateWorkflow ingests GreenNode certificates.
func GreenNodeLoadBalancerCertificateWorkflow(ctx workflow.Context, params GreenNodeLoadBalancerCertificateWorkflowParams) (*GreenNodeLoadBalancerCertificateWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeLoadBalancerCertificateWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestLoadBalancerCertificatesResult
	err := workflow.ExecuteActivity(activityCtx, IngestLoadBalancerCertificatesActivity, IngestLoadBalancerCertificatesParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest certificates", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeLoadBalancerCertificateWorkflow", "certificateCount", result.CertificateCount)

	return &GreenNodeLoadBalancerCertificateWorkflowResult{
		CertificateCount: result.CertificateCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
