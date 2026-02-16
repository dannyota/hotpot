package table

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *ent.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context, projectID string) (*Client, error) {
	var opts []option.ClientOption
	if credJSON := a.configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}
	opts = append(opts, option.WithHTTPClient(&http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, nil),
	}))
	return NewClient(ctx, projectID, opts...)
}

// IngestBigQueryTablesParams contains parameters for the ingest activity.
type IngestBigQueryTablesParams struct {
	ProjectID  string
	DatasetIDs []string
}

// IngestBigQueryTablesResult contains the result of the ingest activity.
type IngestBigQueryTablesResult struct {
	ProjectID      string
	TableCount     int
	DurationMillis int64
}

// IngestBigQueryTablesActivity is the activity function reference for workflow registration.
var IngestBigQueryTablesActivity = (*Activities).IngestBigQueryTables

// IngestBigQueryTables is a Temporal activity that ingests BigQuery tables.
func (a *Activities) IngestBigQueryTables(ctx context.Context, params IngestBigQueryTablesParams) (*IngestBigQueryTablesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting BigQuery table ingestion",
		"projectID", params.ProjectID,
		"datasetCount", len(params.DatasetIDs),
	)

	client, err := a.createClient(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID:  params.ProjectID,
		DatasetIDs: params.DatasetIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest BigQuery tables: %w", err)
	}

	if err := service.DeleteStaleTables(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale BigQuery tables", "error", err)
	}

	logger.Info("Completed BigQuery table ingestion",
		"projectID", params.ProjectID,
		"tableCount", result.TableCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestBigQueryTablesResult{
		ProjectID:      result.ProjectID,
		TableCount:     result.TableCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
