package securitycenter

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/securitycenter/finding"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/securitycenter/notificationconfig"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/securitycenter/source"
)

// GCPSecurityCenterWorkflowParams contains parameters for the SCC workflow.
type GCPSecurityCenterWorkflowParams struct {
}

// GCPSecurityCenterWorkflowResult contains the result of the SCC workflow.
type GCPSecurityCenterWorkflowResult struct {
	SourceCount             int
	FindingCount            int
	NotificationConfigCount int
}

// GCPSecurityCenterWorkflow ingests all Security Command Center resources.
// Executes source workflow first, then findings (findings reference sources).
func GCPSecurityCenterWorkflow(ctx workflow.Context, params GCPSecurityCenterWorkflowParams) (*GCPSecurityCenterWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSecurityCenterWorkflow")

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 60 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPSecurityCenterWorkflowResult{}

	// Phase 1: Ingest sources first (findings reference sources)
	var sourceResult source.GCPSecurityCenterSourceWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, source.GCPSecurityCenterSourceWorkflow,
		source.GCPSecurityCenterSourceWorkflowParams{}).Get(ctx, &sourceResult)
	if err != nil {
		logger.Error("Failed to ingest SCC sources", "error", err)
		return nil, err
	}
	result.SourceCount = sourceResult.SourceCount

	// Phase 2: Ingest findings (depends on sources being in DB)
	var findingResult finding.GCPSecurityCenterFindingWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, finding.GCPSecurityCenterFindingWorkflow,
		finding.GCPSecurityCenterFindingWorkflowParams{}).Get(ctx, &findingResult)
	if err != nil {
		logger.Error("Failed to ingest SCC findings", "error", err)
		return nil, err
	}
	result.FindingCount = findingResult.FindingCount

	// Phase 3: Ingest notification configs (independent of sources/findings)
	var notificationConfigResult notificationconfig.GCPSecurityCenterNotificationConfigWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, notificationconfig.GCPSecurityCenterNotificationConfigWorkflow,
		notificationconfig.GCPSecurityCenterNotificationConfigWorkflowParams{}).Get(ctx, &notificationConfigResult)
	if err != nil {
		logger.Error("Failed to ingest SCC notification configs", "error", err)
		return nil, err
	}
	result.NotificationConfigCount = notificationConfigResult.NotificationConfigCount

	logger.Info("Completed GCPSecurityCenterWorkflow",
		"sourceCount", result.SourceCount,
		"findingCount", result.FindingCount,
		"notificationConfigCount", result.NotificationConfigCount,
	)

	return result, nil
}
