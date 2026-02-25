package securitycenter

import (
	"go.temporal.io/sdk/worker"

	"entgo.io/ent/dialect"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/securitycenter/finding"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/securitycenter/notificationconfig"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/securitycenter/source"
	entresourcemanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/resourcemanager"
	entsecuritycenter "github.com/dannyota/hotpot/pkg/storage/ent/gcp/securitycenter"
)

// Register registers all Security Command Center activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entsecuritycenter.NewClient(entsecuritycenter.Driver(driver), entsecuritycenter.AlternateSchema(entsecuritycenter.DefaultSchemaConfig()))
	rmClient := entresourcemanager.NewClient(entresourcemanager.Driver(driver), entresourcemanager.AlternateSchema(entresourcemanager.DefaultSchemaConfig()))
	source.Register(w, configService, entClient, rmClient, limiter)
	finding.Register(w, configService, entClient, rmClient, limiter)
	notificationconfig.Register(w, configService, entClient, rmClient, limiter)

	w.RegisterWorkflow(GCPSecurityCenterWorkflow)
}
