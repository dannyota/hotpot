package organization

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

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
func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	var opts []option.ClientOption
	if credJSON := a.configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}
	opts = append(opts, option.WithGRPCDialOption(
		grpc.WithUnaryInterceptor(ratelimit.UnaryInterceptor(a.limiter)),
	))
	return NewClient(ctx, opts...)
}

// IngestOrganizationsParams contains parameters for the ingest activity.
type IngestOrganizationsParams struct {
}

// IngestOrganizationsResult contains the result of the ingest activity.
type IngestOrganizationsResult struct {
	OrganizationCount int
	DurationMillis    int64
}

// IngestOrganizationsActivity is the activity function reference for workflow registration.
var IngestOrganizationsActivity = (*Activities).IngestOrganizations

// IngestOrganizations is a Temporal activity that discovers and ingests all accessible GCP organizations.
func (a *Activities) IngestOrganizations(ctx context.Context, params IngestOrganizationsParams) (*IngestOrganizationsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP organization discovery")

	// Create client for this activity
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest organizations: %w", err)
	}

	// Delete stale organizations
	if err := service.DeleteStaleOrganizations(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale organizations", "error", err)
	}

	logger.Info("Completed GCP organization discovery",
		"organizationCount", result.OrganizationCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestOrganizationsResult{
		OrganizationCount: result.OrganizationCount,
		DurationMillis:    result.DurationMillis,
	}, nil
}
