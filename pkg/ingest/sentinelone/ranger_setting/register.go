package ranger_setting

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
)

// Register registers ranger setting activities and workflows with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *ents1.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)

	w.RegisterActivity(activities.IngestS1RangerSettings)

	w.RegisterWorkflow(S1RangerSettingWorkflow)
}
