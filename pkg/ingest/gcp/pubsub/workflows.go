package pubsub

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/pubsub/subscription"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/pubsub/topic"
)

// GCPPubSubWorkflowParams contains parameters for the Pub/Sub workflow.
type GCPPubSubWorkflowParams struct {
	ProjectID string
}

// GCPPubSubWorkflowResult contains the result of the Pub/Sub workflow.
type GCPPubSubWorkflowResult struct {
	ProjectID         string
	TopicCount        int
	SubscriptionCount int
}

// GCPPubSubWorkflow ingests all Pub/Sub resources for a single project.
// Topics and subscriptions are independent and run in parallel.
func GCPPubSubWorkflow(ctx workflow.Context, params GCPPubSubWorkflowParams) (*GCPPubSubWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPPubSubWorkflow", "projectID", params.ProjectID)

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

	result := &GCPPubSubWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Execute topic and subscription workflows in parallel (no dependency)
	topicFuture := workflow.ExecuteChildWorkflow(childCtx, topic.GCPPubSubTopicWorkflow,
		topic.GCPPubSubTopicWorkflowParams{ProjectID: params.ProjectID})

	subscriptionFuture := workflow.ExecuteChildWorkflow(childCtx, subscription.GCPPubSubSubscriptionWorkflow,
		subscription.GCPPubSubSubscriptionWorkflowParams{ProjectID: params.ProjectID})

	// Collect topic result
	var topicResult topic.GCPPubSubTopicWorkflowResult
	if err := topicFuture.Get(ctx, &topicResult); err != nil {
		logger.Error("Failed to ingest Pub/Sub topics", "error", err)
		return nil, err
	}
	result.TopicCount = topicResult.TopicCount

	// Collect subscription result
	var subscriptionResult subscription.GCPPubSubSubscriptionWorkflowResult
	if err := subscriptionFuture.Get(ctx, &subscriptionResult); err != nil {
		logger.Error("Failed to ingest Pub/Sub subscriptions", "error", err)
		return nil, err
	}
	result.SubscriptionCount = subscriptionResult.SubscriptionCount

	logger.Info("Completed GCPPubSubWorkflow",
		"projectID", params.ProjectID,
		"topicCount", result.TopicCount,
		"subscriptionCount", result.SubscriptionCount,
	)

	return result, nil
}
