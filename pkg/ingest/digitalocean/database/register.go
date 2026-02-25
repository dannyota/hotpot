package database

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entdo "github.com/dannyota/hotpot/pkg/storage/ent/do"
)

// Register registers Database activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entdo.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestDODatabases)
	w.RegisterActivity(activities.IngestDODatabaseChildren)

	w.RegisterWorkflow(DODatabaseWorkflow)
}
