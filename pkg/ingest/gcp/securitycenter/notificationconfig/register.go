package notificationconfig

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
	entsecuritycenter "danny.vn/hotpot/pkg/storage/ent/gcp/securitycenter"
)

// Register registers all SCC notification config activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entsecuritycenter.Client, rmClient *entresourcemanager.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, rmClient, limiter)
	w.RegisterActivity(activities.IngestNotificationConfigs)
	w.RegisterWorkflow(GCPSecurityCenterNotificationConfigWorkflow)
}
