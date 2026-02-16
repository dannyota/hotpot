package aws

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"go.temporal.io/sdk/activity"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Activities holds dependencies for AWS-level Temporal activities.
type Activities struct {
	configService *config.Service
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		limiter:       limiter,
	}
}

// DiscoverRegionsParams contains parameters for the region discovery activity.
type DiscoverRegionsParams struct{}

// DiscoverRegionsResult contains the result of region discovery.
type DiscoverRegionsResult struct {
	Regions []string
}

// DiscoverRegionsActivity is the activity function reference for workflow registration.
var DiscoverRegionsActivity = (*Activities).DiscoverRegions

// DiscoverRegions discovers all enabled AWS regions for the account.
// If config Regions is set, filters to only those regions.
func (a *Activities) DiscoverRegions(ctx context.Context, _ DiscoverRegionsParams) (*DiscoverRegionsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Discovering AWS regions")

	cfg, err := a.loadAWSConfig(ctx, "us-east-1")
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	client := awsec2.NewFromConfig(cfg)
	output, err := client.DescribeRegions(ctx, &awsec2.DescribeRegionsInput{
		AllRegions: aws.Bool(false), // Only enabled regions
	})
	if err != nil {
		return nil, fmt.Errorf("describe regions: %w", err)
	}

	var regions []string
	for _, r := range output.Regions {
		if r.RegionName != nil {
			regions = append(regions, *r.RegionName)
		}
	}

	// Filter by configured regions if set
	configRegions := a.configService.AWSRegions()
	if len(configRegions) > 0 {
		allowed := make(map[string]bool, len(configRegions))
		for _, r := range configRegions {
			allowed[r] = true
		}
		var filtered []string
		for _, r := range regions {
			if allowed[r] {
				filtered = append(filtered, r)
			}
		}
		regions = filtered
	}

	logger.Info("Discovered AWS regions", "count", len(regions))
	return &DiscoverRegionsResult{Regions: regions}, nil
}

// loadAWSConfig builds an AWS SDK config with credentials and rate-limited HTTP transport.
func (a *Activities) loadAWSConfig(ctx context.Context, region string) (aws.Config, error) {
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

	return awsconfig.LoadDefaultConfig(ctx, opts...)
}
