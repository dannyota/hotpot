package gcplogging

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/temporalerr"
	"danny.vn/hotpot/pkg/ingest/accesslog"
	entaccesslog "danny.vn/hotpot/pkg/storage/ent/accesslog"
)

// Activities holds dependencies for BigQuery Log Analytics traffic activities.
type Activities struct {
	configService *config.Service
	entClient     *entaccesslog.Client
}

// NewActivities creates an Activities instance.
func NewActivities(configService *config.Service, entClient *entaccesslog.Client) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
	}
}

// IngestTrafficCountsActivity function reference for Temporal registration.
var IngestTrafficCountsActivity = (*Activities).IngestTrafficCounts

// IngestTrafficCounts queries BigQuery Log Analytics and stores aggregated traffic counts.
func (a *Activities) IngestTrafficCounts(ctx context.Context, params accesslog.ServiceWorkflowParams) (*accesslog.ServiceWorkflowResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Ingesting traffic counts",
		"name", params.Name,
		"projectID", params.ProjectID)

	// Look up per-source credentials from config.
	var creds []byte
	found := false
	for _, src := range a.configService.AccessLogSources() {
		if src.Name == params.Name {
			creds = src.CredentialsJSON
			found = true
			break
		}
	}
	if !found {
		return nil, temporalerr.MaybeNonRetryable(
			fmt.Errorf("source %q not found in accesslog config", params.Name))
	}

	bqClient, err := NewBQClient(ctx, creds, params.ProjectID, params.BigQueryTable, params.BQFilter)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create bigquery client: %w", err))
	}
	defer bqClient.Close()

	svc := NewService(bqClient, a.entClient)
	result, err := svc.Ingest(ctx, IngestParams{
		Name:                    params.Name,
		SourceType:              params.SourceType,
		Role:                    params.Role,
		BigQueryTable:           params.BigQueryTable,
		BQFilter:                params.BQFilter,
		FieldMapping:            params.FieldMapping,
		IntervalMinutes:         params.IntervalMinutes,
		BackfillDays:            params.BackfillDays,
		BackfillIntervalMinutes: params.BackfillIntervalMinutes,
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(err)
	}

	logger.Info("Traffic ingestion complete",
		"name", result.Name,
		"windows", result.WindowsIngested,
		"counts", result.CountsCreated)

	return &accesslog.ServiceWorkflowResult{
		Name:   result.Name,
		Counts: result.CountsCreated,
	}, nil
}
