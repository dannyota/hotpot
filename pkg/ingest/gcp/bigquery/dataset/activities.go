package dataset

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/gcpauth"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entbigquery "danny.vn/hotpot/pkg/storage/ent/gcp/bigquery"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entbigquery.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entbigquery.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context, projectID string) (*Client, error) {
	httpClient, err := gcpauth.NewHTTPClient(ctx, a.configService.GCPCredentialsJSON(), a.limiter)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, projectID, option.WithHTTPClient(httpClient))
}

// IngestBigQueryDatasetsParams contains parameters for the ingest activity.
type IngestBigQueryDatasetsParams struct {
	ProjectID string
}

// IngestBigQueryDatasetsResult contains the result of the ingest activity.
type IngestBigQueryDatasetsResult struct {
	ProjectID      string
	DatasetCount   int
	DatasetIDs     []string
	DurationMillis int64
}

// IngestBigQueryDatasetsActivity is the activity function reference for workflow registration.
var IngestBigQueryDatasetsActivity = (*Activities).IngestBigQueryDatasets

// IngestBigQueryDatasets is a Temporal activity that ingests BigQuery datasets.
func (a *Activities) IngestBigQueryDatasets(ctx context.Context, params IngestBigQueryDatasetsParams) (*IngestBigQueryDatasetsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting BigQuery dataset ingestion",
		"projectID", params.ProjectID,
	)

	client, err := a.createClient(ctx, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest BigQuery datasets: %w", err))
	}

	if err := service.DeleteStaleDatasets(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale BigQuery datasets", "error", err)
	}

	logger.Info("Completed BigQuery dataset ingestion",
		"projectID", params.ProjectID,
		"datasetCount", result.DatasetCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestBigQueryDatasetsResult{
		ProjectID:      result.ProjectID,
		DatasetCount:   result.DatasetCount,
		DatasetIDs:     result.DatasetIDs,
		DurationMillis: result.DurationMillis,
	}, nil
}
