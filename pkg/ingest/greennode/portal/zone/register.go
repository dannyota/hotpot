package zone

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entportal "github.com/dannyota/hotpot/pkg/storage/ent/greennode/portal"
)

// Register registers zone workflows and activities with the Temporal worker.
func Register(w worker.Worker, configService *config.Service, entClient *entportal.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, iamAuth, limiter)
	w.RegisterActivity(activities.IngestPortalZones)
	w.RegisterWorkflow(GreenNodePortalZoneWorkflow)
}
