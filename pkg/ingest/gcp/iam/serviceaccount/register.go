package serviceaccount

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
)

func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, db, limiter)
	w.RegisterActivity(activities.IngestIAMServiceAccounts)
	w.RegisterWorkflow(GCPIAMServiceAccountWorkflow)
}
