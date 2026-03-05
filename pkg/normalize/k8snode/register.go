package k8snode

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	entk8snode "github.com/dannyota/hotpot/pkg/storage/ent/k8snode"
)

// Register wires k8s node normalize activities and workflow to the worker.
// Providers are passed in to avoid import cycles (sub-packages import k8snode).
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, db *sql.DB, providers []Provider) {
	entClient := entk8snode.NewClient(
		entk8snode.Driver(driver),
		entk8snode.AlternateSchema(entk8snode.DefaultSchemaConfig()),
	)

	activities := NewActivities(configService, entClient, db, providers)
	w.RegisterActivity(activities.NormalizeK8sNodeProvider)
	w.RegisterActivity(activities.MergeK8sNodes)
	w.RegisterWorkflow(NormalizeK8sNodesWorkflow)
}
