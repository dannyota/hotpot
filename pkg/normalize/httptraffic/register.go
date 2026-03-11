package httptraffic

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/geoip"
	"danny.vn/hotpot/pkg/base/matchrule"
	enthttptraffic "danny.vn/hotpot/pkg/storage/ent/httptraffic"
)

// Register wires HTTP traffic normalize activities and workflow to the worker.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, db *sql.DB) {
	entClient := enthttptraffic.NewClient(
		enthttptraffic.Driver(driver),
		enthttptraffic.AlternateSchema(enthttptraffic.DefaultSchemaConfig()),
	)

	geoipLookup := geoip.NewLookup(configService.GeoIPCityPath(), configService.GeoIPASNPath())

	matchRules := matchrule.NewService(db)
	activities := NewActivities(configService, entClient, db, geoipLookup, matchRules)
	w.RegisterActivity(activities.NormalizeTraffic)
	w.RegisterActivity(activities.NormalizeUserAgents)
	w.RegisterActivity(activities.NormalizeClientIPs)
	w.RegisterWorkflow(NormalizeHttptrafficWorkflow)
}
