package greennode

import (
	"context"
	"fmt"
	"log"
	"time"

	"danny.vn/greennode"
	"danny.vn/greennode/auth"
	"danny.vn/greennode/option"
	portalv1 "danny.vn/greennode/services/portal/v1"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
)

// Activities holds dependencies for top-level GreenNode activities.
type Activities struct {
	configService *config.Service
	limiter       ratelimit.Limiter
	iamAuth       *auth.IAMUserAuth // shared across activities for token caching
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, limiter ratelimit.Limiter) *Activities {
	a := &Activities{configService: configService, limiter: limiter}

	if username := configService.GreenNodeUsername(); username != "" {
		a.iamAuth = &auth.IAMUserAuth{
			RootEmail: configService.GreenNodeRootEmail(),
			Username:  username,
			Password:  configService.GreenNodePassword(),
		}
		if totpSecret := configService.GreenNodeTOTPSecret(); totpSecret != "" {
			a.iamAuth.TOTP = &auth.SecretTOTP{Secret: totpSecret}
		}
	}

	return a
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
	return &DiscoverRegionsResult{
		Regions: a.configService.GreenNodeRegions(),
	}, nil
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
	log.Printf("[DiscoverProjects] start region=%s", params.Region)

	// If project_id is explicitly configured, use it directly.
	if pid := a.configService.GreenNodeProjectID(); pid != "" {
		log.Printf("[DiscoverProjects] using configured project_id=%s", pid)
		return &DiscoverProjectsResult{ProjectIDs: []string{pid}}, nil
	}

	// Discover projects via Portal V1 API (does not require project_id in URL).
	cfg := greennode.Config{
		Region: params.Region,
	}

	if a.iamAuth != nil {
		log.Printf("[DiscoverProjects] using IAM auth (username=%s)", a.iamAuth.Username)
		cfg.IAMAuth = a.iamAuth
	} else {
		log.Printf("[DiscoverProjects] using service account auth")
		cfg.ClientID = a.configService.GreenNodeClientID()
		cfg.ClientSecret = a.configService.GreenNodeClientSecret()
	}

	log.Printf("[DiscoverProjects] creating SDK client...")
	step := time.Now()
	sdk, err := greennode.NewClient(ctx, cfg,
		option.WithTransport(ratelimit.NewRateLimitedTransport(a.limiter, nil)),
	)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create greennode client: %w", err))
	}
	log.Printf("[DiscoverProjects] SDK client created (%v)", time.Since(step))

	log.Printf("[DiscoverProjects] calling ListProjects...")
	step = time.Now()
	result, err := sdk.PortalV1.ListProjects(ctx, portalv1.NewListProjectsRequest())
	if err != nil {
		log.Printf("[DiscoverProjects] ListProjects failed (%v): %v", time.Since(step), err)
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("list projects: %w", err))
	}
	log.Printf("[DiscoverProjects] ListProjects OK (%v) %d projects", time.Since(step), len(result.Items))

	ids := make([]string, 0, len(result.Items))
	for _, p := range result.Items {
		ids = append(ids, p.ProjectID)
	}

	return &DiscoverProjectsResult{ProjectIDs: ids}, nil
}
