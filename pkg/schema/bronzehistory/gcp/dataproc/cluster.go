package dataproc

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPDataprocCluster stores historical snapshots of GCP Dataproc clusters.
type BronzeHistoryGCPDataprocCluster struct {
	ent.Schema
}

func (BronzeHistoryGCPDataprocCluster) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPDataprocCluster) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze cluster by resource_id"),

		// All cluster fields
		field.String("cluster_name").
			NotEmpty(),
		field.String("cluster_uuid").
			Optional(),

		// JSONB fields
		field.JSON("config_json", json.RawMessage{}).
			Optional(),
		field.JSON("status_json", json.RawMessage{}).
			Optional(),
		field.JSON("status_history_json", json.RawMessage{}).
			Optional(),
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),
		field.JSON("metrics_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			NotEmpty().
			Comment("Region where the cluster runs"),
	}
}

func (BronzeHistoryGCPDataprocCluster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPDataprocCluster) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_dataproc_clusters_history"},
	}
}
