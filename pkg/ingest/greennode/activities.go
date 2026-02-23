package greennode

import (
	"context"

	"github.com/dannyota/hotpot/pkg/base/config"
)

// Activities holds dependencies for top-level GreenNode activities.
type Activities struct {
	configService *config.Service
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service) *Activities {
	return &Activities{configService: configService}
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
