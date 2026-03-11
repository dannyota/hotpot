package ubuntu

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entreference "danny.vn/hotpot/pkg/storage/ent/reference"
)

// Register registers Ubuntu package activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entreference.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestUbuntuFeed)

	w.RegisterWorkflow(UbuntuPackagesWorkflow)
}
