package logging

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/logbucket"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/logexclusion"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/logmetric"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging/sink"
)

// GCPLoggingWorkflowParams contains parameters for the logging workflow.
type GCPLoggingWorkflowParams struct {
	ProjectID string
}

// GCPLoggingWorkflowResult contains the result of the logging workflow.
type GCPLoggingWorkflowResult struct {
	ProjectID      string
	SinkCount      int
	BucketCount    int
	LogMetricCount int
	ExclusionCount int
}

// GCPLoggingWorkflow ingests all GCP Cloud Logging resources for a single project.
func GCPLoggingWorkflow(ctx workflow.Context, params GCPLoggingWorkflowParams) (*GCPLoggingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPLoggingWorkflow", "projectID", params.ProjectID)

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

	result := &GCPLoggingWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Execute sink workflow
	var sinkResult sink.GCPLoggingSinkWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, sink.GCPLoggingSinkWorkflow,
		sink.GCPLoggingSinkWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &sinkResult)
	if err != nil {
		logger.Error("Failed to ingest sinks", "error", err)
		return nil, err
	}
	result.SinkCount = sinkResult.SinkCount

	// Execute bucket workflow
	var bucketResult logbucket.GCPLoggingBucketWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, logbucket.GCPLoggingBucketWorkflow,
		logbucket.GCPLoggingBucketWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &bucketResult)
	if err != nil {
		logger.Error("Failed to ingest buckets", "error", err)
		return nil, err
	}
	result.BucketCount = bucketResult.BucketCount

	// Execute log metric workflow
	var logMetricResult logmetric.GCPLoggingLogMetricWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, logmetric.GCPLoggingLogMetricWorkflow,
		logmetric.GCPLoggingLogMetricWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &logMetricResult)
	if err != nil {
		logger.Error("Failed to ingest log metrics", "error", err)
		return nil, err
	}
	result.LogMetricCount = logMetricResult.LogMetricCount

	// Execute log exclusion workflow
	var logExclusionResult logexclusion.GCPLoggingLogExclusionWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, logexclusion.GCPLoggingLogExclusionWorkflow,
		logexclusion.GCPLoggingLogExclusionWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &logExclusionResult)
	if err != nil {
		logger.Error("Failed to ingest log exclusions", "error", err)
		return nil, err
	}
	result.ExclusionCount = logExclusionResult.ExclusionCount

	logger.Info("Completed GCPLoggingWorkflow",
		"projectID", params.ProjectID,
		"sinkCount", result.SinkCount,
		"bucketCount", result.BucketCount,
		"logMetricCount", result.LogMetricCount,
		"exclusionCount", result.ExclusionCount,
	)

	return result, nil
}
