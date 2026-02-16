package domain

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DODomainWorkflowResult contains the result of the Domain workflow.
type DODomainWorkflowResult struct {
	DomainCount    int
	RecordCount    int
	DurationMillis int64
}

// DODomainWorkflow ingests DigitalOcean Domains and their Records.
func DODomainWorkflow(ctx workflow.Context) (*DODomainWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DODomainWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	// Step 1: Ingest domains
	var domainsResult IngestDODomainsResult
	err := workflow.ExecuteActivity(activityCtx, IngestDODomainsActivity).Get(ctx, &domainsResult)
	if err != nil {
		logger.Error("Failed to ingest domains", "error", err)
		return nil, err
	}

	// Step 2: Ingest domain records using the domain names from step 1
	var recordsResult IngestDODomainRecordsResult
	if len(domainsResult.DomainNames) > 0 {
		err = workflow.ExecuteActivity(activityCtx, IngestDODomainRecordsActivity, IngestDODomainRecordsInput{
			DomainNames: domainsResult.DomainNames,
		}).Get(ctx, &recordsResult)
		if err != nil {
			logger.Error("Failed to ingest domain records", "error", err)
			return nil, err
		}
	}

	logger.Info("Completed DODomainWorkflow",
		"domainCount", domainsResult.DomainCount,
		"recordCount", recordsResult.RecordCount,
	)

	return &DODomainWorkflowResult{
		DomainCount:    domainsResult.DomainCount,
		RecordCount:    recordsResult.RecordCount,
		DurationMillis: domainsResult.DurationMillis + recordsResult.DurationMillis,
	}, nil
}
