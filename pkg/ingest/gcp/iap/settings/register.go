package settings

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entiap "danny.vn/hotpot/pkg/storage/ent/gcp/iap"
)

// Register registers all IAP settings activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entiap.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestIAPSettings)
	w.RegisterWorkflow(GCPIAPSettingsWorkflow)
}
