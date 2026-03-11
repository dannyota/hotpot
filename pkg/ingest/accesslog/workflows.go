package accesslog

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
	"danny.vn/hotpot/pkg/ingest"
)

// AccessLogWorkflowResult holds the result of the access log workflow.
type AccessLogWorkflowResult struct {
	SourcesDiscovered int
	SourceResults     []SourceResult
}

// SourceResult holds the ingestion result for a single log source.
type SourceResult struct {
	Name  string
	Error string
	Counts int
}

// serviceWorkflowFunc is the func signature for source-type service workflows.
type serviceWorkflowFunc = func(workflow.Context, ServiceWorkflowParams) (*ServiceWorkflowResult, error)

// ServiceWorkflowParams holds parameters passed to source-type service workflows.
type ServiceWorkflowParams struct {
	Name            string
	SourceType      string
	Role            string
	ProjectID       string
	BigQueryTable   string
	BQFilter        string
	FieldMapping    map[string]string
	IntervalMinutes int

	// Backfill settings for first run (no cursor).
	BackfillDays            int
	BackfillIntervalMinutes int
}

// ServiceWorkflowResult holds the result from a source-type service workflow.
type ServiceWorkflowResult struct {
	Name   string
	Counts int
}

// AccessLogWorkflow discovers log sources and dispatches child workflows by source type.
func AccessLogWorkflow(ctx workflow.Context) (*AccessLogWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting AccessLogWorkflow")

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

	// Phase 1: Read log sources from config.
	var discoverResult DiscoverLogSourcesResult
	if err := workflow.ExecuteActivity(activityCtx, DiscoverLogSourcesActivity).
		Get(ctx, &discoverResult); err != nil {
		logger.Error("Failed to discover log sources", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	if len(discoverResult.Sources) == 0 {
		logger.Info("No log sources configured")
		return &AccessLogWorkflowResult{}, nil
	}

	// Build service lookup by source type.
	serviceLookup := make(map[string]ingest.ServiceRegistration)
	for _, svc := range ingest.Services("accesslog") {
		serviceLookup[svc.Name] = svc
	}

	// Phase 2: Launch child workflows per source.
	// Use a longer timeout when backfill is configured (first run may process many windows).
	childTimeout := 30 * time.Minute
	for _, src := range discoverResult.Sources {
		if src.BackfillDays > 0 {
			childTimeout = 2 * time.Hour
			break
		}
	}
	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: childTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	type sourceCall struct {
		name   string
		future workflow.ChildWorkflowFuture
	}
	var calls []sourceCall

	for _, src := range discoverResult.Sources {
		svc, ok := serviceLookup[src.SourceType]
		if !ok {
			logger.Warn("No service registered for source type",
				"sourceType", src.SourceType, "name", src.Name)
			continue
		}

		params := ServiceWorkflowParams{
			Name:                    src.Name,
			SourceType:              src.SourceType,
			Role:                    src.Role,
			ProjectID:               src.ProjectID,
			BigQueryTable:           src.BigQueryTable,
			BQFilter:                src.BQFilter,
			FieldMapping:            src.FieldMapping,
			IntervalMinutes:         src.IntervalMinutes,
			BackfillDays:            src.BackfillDays,
			BackfillIntervalMinutes: src.BackfillIntervalMinutes,
		}
		f := workflow.ExecuteChildWorkflow(childCtx, svc.Workflow, params)
		calls = append(calls, sourceCall{name: src.Name, future: f})
	}

	// Phase 3: Collect results.
	result := &AccessLogWorkflowResult{
		SourcesDiscovered: len(discoverResult.Sources),
		SourceResults:     make([]SourceResult, 0, len(calls)),
	}

	for _, c := range calls {
		var svcResult ServiceWorkflowResult
		if err := c.future.Get(ctx, &svcResult); err != nil {
			logger.Error("Failed to ingest source", "name", c.name, "error", err)
			result.SourceResults = append(result.SourceResults, SourceResult{
				Name:  c.name,
				Error: err.Error(),
			})
		} else {
			result.SourceResults = append(result.SourceResults, SourceResult{
				Name:   c.name,
				Counts: svcResult.Counts,
			})
		}
	}

	// Phase 4: Cleanup stale bronze data.
	var cleanupResult CleanupStaleBronzeResult
	if err := workflow.ExecuteActivity(activityCtx, CleanupStaleBronzeActivity).
		Get(ctx, &cleanupResult); err != nil {
		logger.Error("CleanupStaleBronze failed", "error", err)
		// Non-fatal: log but don't fail the workflow.
	} else {
		logger.Info("CleanupStaleBronze done",
			"bronzeCountsDeleted", cleanupResult.BronzeCountsDeleted)
	}

	logger.Info("Completed AccessLogWorkflow",
		"sourcesDiscovered", result.SourcesDiscovered,
		"sourcesProcessed", len(result.SourceResults))

	return result, nil
}
