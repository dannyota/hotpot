package httpmonitor

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/matchrule"
	enthttpmonitor "danny.vn/hotpot/pkg/storage/ent/httpmonitor"
)

// Register wires httpmonitor anomaly detection activities and workflow to the worker.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, db *sql.DB) {
	entClient := enthttpmonitor.NewClient(
		enthttpmonitor.Driver(driver),
		enthttpmonitor.AlternateSchema(enthttpmonitor.DefaultSchemaConfig()),
	)

	matchRules := matchrule.NewService(db)
	activities := NewActivities(configService, entClient, db, matchRules)
	w.RegisterActivity(activities.DetectRateAnomalies)
	w.RegisterActivity(activities.DetectErrorBursts)
	w.RegisterActivity(activities.DetectSuspiciousPatterns)
	w.RegisterActivity(activities.DetectMethodMismatch)
	w.RegisterActivity(activities.DetectUserAgentAnomalies)
	w.RegisterActivity(activities.DetectClientIPAnomalies)
	w.RegisterActivity(activities.DetectASNAnomalies)
	w.RegisterActivity(activities.DetectNewEndpoints)
	w.RegisterActivity(activities.DetectAuthAnomalies)
	w.RegisterActivity(activities.CleanupStale)
	w.RegisterWorkflow(HttpMonitorAnomalyWorkflow)
}
