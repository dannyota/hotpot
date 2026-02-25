package vault

import (
	"context"

	"go.temporal.io/sdk/activity"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Activities holds dependencies for Vault provider-level Temporal activities.
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

// ListVaultInstancesResult contains the result of listing vault instances.
type ListVaultInstancesResult struct {
	VaultNames []string
}

// ListVaultInstancesActivity is the activity function reference for workflow registration.
var ListVaultInstancesActivity = (*Activities).ListVaultInstances

// ListVaultInstances reads vault instance names from config.
func (a *Activities) ListVaultInstances(ctx context.Context) (*ListVaultInstancesResult, error) {
	logger := activity.GetLogger(ctx)

	instances := a.configService.VaultInstances()
	names := make([]string, 0, len(instances))
	for _, inst := range instances {
		names = append(names, inst.Name)
	}

	logger.Info("Listed Vault instances", "count", len(names))

	return &ListVaultInstancesResult{
		VaultNames: names,
	}, nil
}
