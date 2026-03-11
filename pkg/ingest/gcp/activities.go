package gcp

import (
	"context"
	"fmt"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	serviceusage "cloud.google.com/go/serviceusage/apiv1"
	"cloud.google.com/go/serviceusage/apiv1/serviceusagepb"
	"go.temporal.io/sdk/activity"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"danny.vn/hotpot/pkg/base/temporalerr"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
)

// Activities holds dependencies for GCP provider-level Temporal activities.
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

// DiscoverProjectsParams contains parameters for the project discovery activity.
type DiscoverProjectsParams struct{}

// DiscoverProjectsResult contains the result of project discovery.
type DiscoverProjectsResult struct {
	ProjectIDs []string
}

// DiscoverProjectsActivity is the activity function reference for workflow registration.
var DiscoverProjectsActivity = (*Activities).DiscoverProjects

// DiscoverProjects discovers all active GCP projects accessible by the service account.
// This is a lightweight discovery — full project ingestion (DB writes, orgs, folders, IAM)
// is handled by the resourcemanager service.
func (a *Activities) DiscoverProjects(ctx context.Context, _ DiscoverProjectsParams) (*DiscoverProjectsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Discovering GCP projects")

	var opts []option.ClientOption
	if credJSON := a.configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}
	opts = append(opts, option.WithGRPCDialOption(
		grpc.WithUnaryInterceptor(ratelimit.UnaryInterceptor(a.limiter)),
	))

	client, err := resourcemanager.NewProjectsClient(ctx, opts...)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create projects client: %w", err))
	}
	defer client.Close()

	it := client.SearchProjects(ctx, &resourcemanagerpb.SearchProjectsRequest{})

	var projectIDs []string
	for {
		proj, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("search projects: %w", err))
		}

		if proj.GetState() == resourcemanagerpb.Project_ACTIVE {
			projectIDs = append(projectIDs, proj.GetProjectId())
		}
	}

	logger.Info("Discovered GCP projects", "count", len(projectIDs))
	return &DiscoverProjectsResult{ProjectIDs: projectIDs}, nil
}

// GetConfigQuotaProjectResult contains the configured GCP quota project.
type GetConfigQuotaProjectResult struct {
	QuotaProject string
}

// GetConfigQuotaProjectActivity is the activity function reference.
var GetConfigQuotaProjectActivity = (*Activities).GetConfigQuotaProject

// GetConfigQuotaProject returns the configured GCP quota project from config.
func (a *Activities) GetConfigQuotaProject(ctx context.Context) (*GetConfigQuotaProjectResult, error) {
	return &GetConfigQuotaProjectResult{QuotaProject: a.configService.GCPQuotaProject()}, nil
}

// DiscoverEnabledAPIsParams contains parameters for the enabled API discovery activity.
type DiscoverEnabledAPIsParams struct {
	ProjectID string
}

// DiscoverEnabledAPIsResult contains the result of enabled API discovery.
type DiscoverEnabledAPIsResult struct {
	EnabledAPIs []string
}

// DiscoverEnabledAPIsActivity is the activity function reference for workflow registration.
var DiscoverEnabledAPIsActivity = (*Activities).DiscoverEnabledAPIs

// DiscoverEnabledAPIs lists which GCP APIs are enabled for a project.
// This is a lightweight read-only call — no database writes.
func (a *Activities) DiscoverEnabledAPIs(ctx context.Context, params DiscoverEnabledAPIsParams) (*DiscoverEnabledAPIsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Discovering enabled APIs", "projectID", params.ProjectID)

	var opts []option.ClientOption
	if credJSON := a.configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}
	opts = append(opts, option.WithGRPCDialOption(
		grpc.WithUnaryInterceptor(ratelimit.UnaryInterceptor(a.limiter)),
	))

	client, err := serviceusage.NewClient(ctx, opts...)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create serviceusage client: %w", err))
	}
	defer client.Close()

	it := client.ListServices(ctx, &serviceusagepb.ListServicesRequest{
		Parent: "projects/" + params.ProjectID,
		Filter: "state:ENABLED",
	})

	var apis []string
	for {
		svc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("list enabled services for %s: %w", params.ProjectID, err))
		}
		// svc.GetName() is like "projects/123/services/compute.googleapis.com"
		name := svc.GetName()
		if i := len(name) - 1; i >= 0 {
			for i > 0 && name[i-1] != '/' {
				i--
			}
			apis = append(apis, name[i:])
		}
	}

	logger.Info("Discovered enabled APIs", "projectID", params.ProjectID, "count", len(apis))
	return &DiscoverEnabledAPIsResult{EnabledAPIs: apis}, nil
}
