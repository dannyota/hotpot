package logging

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
	"danny.vn/hotpot/pkg/ingest/gcp/logging/logbucket"
	"danny.vn/hotpot/pkg/ingest/gcp/logging/logexclusion"
	"danny.vn/hotpot/pkg/ingest/gcp/logging/logmetric"
	"danny.vn/hotpot/pkg/ingest/gcp/logging/sink"
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
// All four resources are independent and run in parallel.
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

	// Launch all four resources in parallel (they are independent)
	sinkFuture := workflow.ExecuteChildWorkflow(childCtx, sink.GCPLoggingSinkWorkflow,
		sink.GCPLoggingSinkWorkflowParams{ProjectID: params.ProjectID})

	bucketFuture := workflow.ExecuteChildWorkflow(childCtx, logbucket.GCPLoggingBucketWorkflow,
		logbucket.GCPLoggingBucketWorkflowParams{ProjectID: params.ProjectID})

	logMetricFuture := workflow.ExecuteChildWorkflow(childCtx, logmetric.GCPLoggingLogMetricWorkflow,
		logmetric.GCPLoggingLogMetricWorkflowParams{ProjectID: params.ProjectID})

	logExclusionFuture := workflow.ExecuteChildWorkflow(childCtx, logexclusion.GCPLoggingLogExclusionWorkflow,
		logexclusion.GCPLoggingLogExclusionWorkflowParams{ProjectID: params.ProjectID})

	// Collect results
	var sinkResult sink.GCPLoggingSinkWorkflowResult
	if err := sinkFuture.Get(ctx, &sinkResult); err != nil {
		logger.Error("Failed to ingest sinks", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}
	result.SinkCount = sinkResult.SinkCount

	var bucketResult logbucket.GCPLoggingBucketWorkflowResult
	if err := bucketFuture.Get(ctx, &bucketResult); err != nil {
		logger.Error("Failed to ingest buckets", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}
	result.BucketCount = bucketResult.BucketCount

	var logMetricResult logmetric.GCPLoggingLogMetricWorkflowResult
	if err := logMetricFuture.Get(ctx, &logMetricResult); err != nil {
		logger.Error("Failed to ingest log metrics", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}
	result.LogMetricCount = logMetricResult.LogMetricCount

	var logExclusionResult logexclusion.GCPLoggingLogExclusionWorkflowResult
	if err := logExclusionFuture.Get(ctx, &logExclusionResult); err != nil {
		logger.Error("Failed to ingest log exclusions", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
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
