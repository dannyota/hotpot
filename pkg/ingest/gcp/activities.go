package gcp

import (
	"context"

	"github.com/dannyota/hotpot/pkg/base/config"
)

// Activities holds dependencies for top-level GCP activities.
type Activities struct {
	configService *config.Service
}

// ResolveDisabledServicesActivity is the activity function reference for workflow registration.
var ResolveDisabledServicesActivity = (*Activities).ResolveDisabledServices

// ResolveDisabledServices returns the list of disabled GCP services from config.
func (a *Activities) ResolveDisabledServices(ctx context.Context) ([]string, error) {
	return a.configService.GCPDisabledServices(), nil
}
