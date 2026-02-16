package instance

import (
	"context"
	"fmt"
	"net/http"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"go.temporal.io/sdk/activity"

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

// createClient creates a rate-limited AWS EC2 client for the given region.
func (a *Activities) createClient(ctx context.Context, region string) (*Client, error) {
	var opts []func(*awsconfig.LoadOptions) error

	opts = append(opts, awsconfig.WithRegion(region))

	// Static credentials if configured
	if accessKey := a.configService.AWSAccessKeyID(); accessKey != "" {
		secretKey := a.configService.AWSSecretAccessKey()
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		))
	}

	// Rate-limited HTTP client
	opts = append(opts, awsconfig.WithHTTPClient(&http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, nil),
	}))

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	return NewClient(cfg), nil
}

// IngestEC2InstancesParams contains parameters for the ingest activity.
type IngestEC2InstancesParams struct {
	AccountID string
	Region    string
}

// IngestEC2InstancesResult contains the result of the ingest activity.
type IngestEC2InstancesResult struct {
	AccountID      string
	Region         string
	InstanceCount  int
	DurationMillis int64
}

// IngestEC2InstancesActivity is the activity function reference for workflow registration.
var IngestEC2InstancesActivity = (*Activities).IngestEC2Instances

// IngestEC2Instances is a Temporal activity that ingests AWS EC2 instances.
func (a *Activities) IngestEC2Instances(ctx context.Context, params IngestEC2InstancesParams) (*IngestEC2InstancesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting AWS EC2 instance ingestion",
		"accountID", params.AccountID,
		"region", params.Region,
	)

	client, err := a.createClient(ctx, params.Region)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		AccountID: params.AccountID,
		Region:    params.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest instances: %w", err)
	}

	// Delete stale instances
	if err := service.DeleteStaleInstances(ctx, params.AccountID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale instances", "error", err)
	}

	logger.Info("Completed AWS EC2 instance ingestion",
		"accountID", params.AccountID,
		"region", params.Region,
		"instanceCount", result.InstanceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestEC2InstancesResult{
		AccountID:      result.AccountID,
		Region:         result.Region,
		InstanceCount:  result.InstanceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
