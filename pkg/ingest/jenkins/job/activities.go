package job

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"go.temporal.io/sdk/activity"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entjenkins "github.com/dannyota/hotpot/pkg/storage/ent/jenkins"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entjenkins.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entjenkins.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

func (a *Activities) createClient() *Client {
	transport := ratelimit.NewRateLimitedTransport(a.limiter, nil)

	if !a.configService.JenkinsVerifySSL() {
		transport = ratelimit.NewRateLimitedTransport(a.limiter, &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // user-configured
		})
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(a.configService.JenkinsTimeout()) * time.Second,
	}

	return NewClient(
		a.configService.JenkinsBaseURL(),
		a.configService.JenkinsUsername(),
		a.configService.JenkinsAPIToken(),
		a.configService.JenkinsMaxBuildsPerJob(),
		a.configService.JenkinsExcludeRepos(),
		httpClient,
	)
}

// IngestJenkinsJobsResult contains the result of the ingest activity.
type IngestJenkinsJobsResult struct {
	JobCount       int
	BuildCount     int
	RepoCount      int
	DurationMillis int64
}

// IngestJenkinsJobsActivity is the activity function reference for workflow registration.
var IngestJenkinsJobsActivity = (*Activities).IngestJenkinsJobs

// IngestJenkinsJobs is a Temporal activity that ingests Jenkins jobs, builds, and repos.
func (a *Activities) IngestJenkinsJobs(ctx context.Context) (*IngestJenkinsJobsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Jenkins job ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	since := a.configService.JenkinsSince()

	result, err := service.Ingest(ctx, since, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest jobs: %w", err)
	}

	if err := service.DeleteStale(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale jobs", "error", err)
	}

	logger.Info("Completed Jenkins job ingestion",
		"jobCount", result.JobCount,
		"buildCount", result.BuildCount,
		"repoCount", result.RepoCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestJenkinsJobsResult{
		JobCount:       result.JobCount,
		BuildCount:     result.BuildCount,
		RepoCount:      result.RepoCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
