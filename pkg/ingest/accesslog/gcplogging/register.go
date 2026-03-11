package gcplogging

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	entaccesslog "danny.vn/hotpot/pkg/storage/ent/accesslog"
)

// Register registers BigQuery Log Analytics traffic activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entaccesslog.Client) {
	activities := NewActivities(configService, entClient)
	w.RegisterActivity(activities.IngestTrafficCounts)
	w.RegisterWorkflow(GcpLoggingTrafficWorkflow)
}
