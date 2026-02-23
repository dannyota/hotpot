package greennode

import (
	"context"
	"fmt"

	"danny.vn/greennode"
	"danny.vn/greennode/auth"
	"danny.vn/greennode/option"
	portalv1 "danny.vn/greennode/services/portal/v1"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Activities holds dependencies for top-level GreenNode activities.
type Activities struct {
	configService *config.Service
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, limiter ratelimit.Limiter) *Activities {
	return &Activities{configService: configService, limiter: limiter}
}

// DiscoverRegionsParams contains parameters for region discovery.
type DiscoverRegionsParams struct{}

// DiscoverRegionsResult contains the discovered regions.
type DiscoverRegionsResult struct {
	Regions []string
}

// DiscoverRegionsActivity is the activity function reference for workflow registration.
var DiscoverRegionsActivity = (*Activities).DiscoverRegions

// DiscoverRegions reads configured GreenNode regions from config.
func (a *Activities) DiscoverRegions(ctx context.Context, _ DiscoverRegionsParams) (*DiscoverRegionsResult, error) {
	return &DiscoverRegionsResult{Regions: a.configService.GreenNodeRegions()}, nil
}

// DiscoverProjectsParams contains parameters for project discovery.
type DiscoverProjectsParams struct {
	Region string
}

// DiscoverProjectsResult contains the discovered projects.
type DiscoverProjectsResult struct {
	ProjectIDs []string
}

// DiscoverProjectsActivity is the activity function reference for workflow registration.
var DiscoverProjectsActivity = (*Activities).DiscoverProjects

// DiscoverProjects discovers GreenNode project IDs.
// If project_id is configured, it returns that single value.
// Otherwise, it calls the Portal V1 API to list all accessible projects.
func (a *Activities) DiscoverProjects(ctx context.Context, params DiscoverProjectsParams) (*DiscoverProjectsResult, error) {
	// If project_id is explicitly configured, use it directly.
	if pid := a.configService.GreenNodeProjectID(); pid != "" {
		return &DiscoverProjectsResult{ProjectIDs: []string{pid}}, nil
	}

	// Discover projects via Portal V1 API (does not require project_id in URL).
	cfg := greennode.Config{
		Region: params.Region,
	}

	if username := a.configService.GreenNodeUsername(); username != "" {
		iamAuth := &auth.IAMUserAuth{
			RootEmail: a.configService.GreenNodeRootEmail(),
			Username:  username,
			Password:  a.configService.GreenNodePassword(),
		}
		if totpSecret := a.configService.GreenNodeTOTPSecret(); totpSecret != "" {
			iamAuth.TOTP = &auth.SecretTOTP{Secret: totpSecret}
		}
		cfg.IAMAuth = iamAuth
	} else {
		cfg.ClientID = a.configService.GreenNodeClientID()
		cfg.ClientSecret = a.configService.GreenNodeClientSecret()
	}

	sdk, err := greennode.NewClient(ctx, cfg,
		option.WithTransport(ratelimit.NewRateLimitedTransport(a.limiter, nil)),
	)
	if err != nil {
		return nil, fmt.Errorf("create greennode client: %w", err)
	}

	result, err := sdk.PortalV1.ListProjects(ctx, portalv1.NewListProjectsRequest())
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	ids := make([]string, 0, len(result.Items))
	for _, p := range result.Items {
		ids = append(ids, p.ProjectID)
	}

	return &DiscoverProjectsResult{ProjectIDs: ids}, nil
}
