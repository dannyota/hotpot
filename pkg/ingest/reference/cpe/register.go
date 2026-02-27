package cpe

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entreference "github.com/dannyota/hotpot/pkg/storage/ent/reference"
)

// Register registers CPE activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entreference.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestCPE)

	w.RegisterWorkflow(CPEWorkflow)
}
