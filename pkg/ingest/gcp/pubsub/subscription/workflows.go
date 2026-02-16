package subscription

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPPubSubSubscriptionWorkflowParams contains parameters for the subscription workflow.
type GCPPubSubSubscriptionWorkflowParams struct {
	ProjectID string
}

// GCPPubSubSubscriptionWorkflowResult contains the result of the subscription workflow.
type GCPPubSubSubscriptionWorkflowResult struct {
	ProjectID         string
	SubscriptionCount int
	DurationMillis    int64
}

// GCPPubSubSubscriptionWorkflow ingests Pub/Sub subscriptions for a single project.
func GCPPubSubSubscriptionWorkflow(ctx workflow.Context, params GCPPubSubSubscriptionWorkflowParams) (*GCPPubSubSubscriptionWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPPubSubSubscriptionWorkflow", "projectID", params.ProjectID)

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

	var result IngestPubSubSubscriptionsResult
	err := workflow.ExecuteActivity(activityCtx, IngestPubSubSubscriptionsActivity, IngestPubSubSubscriptionsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Pub/Sub subscriptions", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPPubSubSubscriptionWorkflow",
		"projectID", params.ProjectID,
		"subscriptionCount", result.SubscriptionCount,
	)

	return &GCPPubSubSubscriptionWorkflowResult{
		ProjectID:         result.ProjectID,
		SubscriptionCount: result.SubscriptionCount,
		DurationMillis:    result.DurationMillis,
	}, nil
}
