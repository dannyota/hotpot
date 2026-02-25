package iam

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iam/serviceaccount"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iam/serviceaccountkey"
	"entgo.io/ent/dialect"
	entiam "github.com/dannyota/hotpot/pkg/storage/ent/gcp/iam"
)

// Register registers all IAM activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entiam.NewClient(entiam.Driver(driver), entiam.AlternateSchema(entiam.DefaultSchemaConfig()))
	serviceaccount.Register(w, configService, entClient, limiter)
	serviceaccountkey.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPIAMWorkflow)
}
