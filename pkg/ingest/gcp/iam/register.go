package iam

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/iam/serviceaccount"
	"danny.vn/hotpot/pkg/ingest/gcp/iam/serviceaccountkey"
	"entgo.io/ent/dialect"
	entiam "danny.vn/hotpot/pkg/storage/ent/gcp/iam"
)

// Register registers all IAM activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entiam.NewClient(entiam.Driver(driver), entiam.AlternateSchema(entiam.DefaultSchemaConfig()))
	serviceaccount.Register(w, configService, entClient, limiter)
	serviceaccountkey.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPIAMWorkflow)
}
