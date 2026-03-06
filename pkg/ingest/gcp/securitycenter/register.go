package securitycenter

import (
	"go.temporal.io/sdk/worker"

	"entgo.io/ent/dialect"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/securitycenter/finding"
	"danny.vn/hotpot/pkg/ingest/gcp/securitycenter/notificationconfig"
	"danny.vn/hotpot/pkg/ingest/gcp/securitycenter/source"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
	entsecuritycenter "danny.vn/hotpot/pkg/storage/ent/gcp/securitycenter"
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
