package topic

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPPubSubTopicWorkflowParams contains parameters for the topic workflow.
type GCPPubSubTopicWorkflowParams struct {
	ProjectID string
}

// GCPPubSubTopicWorkflowResult contains the result of the topic workflow.
type GCPPubSubTopicWorkflowResult struct {
	ProjectID      string
	TopicCount     int
	DurationMillis int64
}

// GCPPubSubTopicWorkflow ingests Pub/Sub topics for a single project.
func GCPPubSubTopicWorkflow(ctx workflow.Context, params GCPPubSubTopicWorkflowParams) (*GCPPubSubTopicWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPPubSubTopicWorkflow", "projectID", params.ProjectID)

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

	var result IngestPubSubTopicsResult
	err := workflow.ExecuteActivity(activityCtx, IngestPubSubTopicsActivity, IngestPubSubTopicsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Pub/Sub topics", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPPubSubTopicWorkflow",
		"projectID", params.ProjectID,
		"topicCount", result.TopicCount,
	)

	return &GCPPubSubTopicWorkflowResult{
		ProjectID:      result.ProjectID,
		TopicCount:     result.TopicCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
