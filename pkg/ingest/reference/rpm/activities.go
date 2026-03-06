package rpm

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entreference "danny.vn/hotpot/pkg/storage/ent/reference"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entreference.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entreference.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

func (a *Activities) createClient() *Client {
	httpClient := &http.Client{
		Transport: &userAgentTransport{
			base: ratelimit.NewRateLimitedTransport(a.limiter, nil),
		},
	}
	return NewClient(httpClient)
}

// userAgentTransport wraps an http.RoundTripper to set a DNF-like User-Agent,
// so RPM mirror servers treat us as a normal package manager client.
type userAgentTransport struct {
	base http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", "dnf/4.14.0")
	return t.base.RoundTrip(req)
}

// IngestRPMRepoInput is the input for the per-repo ingest activity.
type IngestRPMRepoInput struct {
	RepoName string
}

// IngestRPMRepoResult contains the result of a single RPM repo ingest activity.
type IngestRPMRepoResult struct {
	RepoName       string
	PackageCount   int
	DurationMillis int64
}

// IngestRPMRepoActivity is the activity function reference for workflow registration.
var IngestRPMRepoActivity = (*Activities).IngestRPMRepo

// IngestRPMRepo downloads and ingests a single RPM repository.
func (a *Activities) IngestRPMRepo(ctx context.Context, input IngestRPMRepoInput) (*IngestRPMRepoResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting RPM repo ingestion", "repo", input.RepoName)

	// Find repo definition
	var repo RepoDef
	var found bool
	for _, r := range Repos {
		if r.Name == input.RepoName {
			repo = r
			found = true
			break
		}
	}
	if !found {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("unknown RPM repo: %s", input.RepoName))
	}

	heartbeat := func(details string) {
		activity.RecordHeartbeat(ctx, details)
	}

	client := a.createClient()
	packages, err := client.DownloadRepo(ctx, repo, heartbeat)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("download RPM repo %s: %w", input.RepoName, err))
	}

	service := NewService(a.entClient)
	result, err := service.IngestRepo(ctx, input.RepoName, packages, heartbeat)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest RPM repo %s: %w", input.RepoName, err))
	}

	logger.Info("Completed RPM repo ingestion",
		"repo", result.RepoName,
		"packageCount", result.PackageCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestRPMRepoResult{
		RepoName:       result.RepoName,
		PackageCount:   result.PackageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
