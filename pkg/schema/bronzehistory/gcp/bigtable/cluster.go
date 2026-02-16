package bigtable

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPBigtableCluster stores historical snapshots of GCP Bigtable clusters.
type BronzeHistoryGCPBigtableCluster struct {
	ent.Schema
}

func (BronzeHistoryGCPBigtableCluster) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPBigtableCluster) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze cluster by resource_id"),

		// All cluster fields
		field.String("location").
			Optional().
			Comment("Zone or region location of the cluster"),
		field.Int32("state").
			Optional().
			Comment("Cluster state: 0=STATE_NOT_KNOWN, 1=READY, 2=CREATING, 3=RESIZING, 4=DISABLED"),
		field.Int32("serve_nodes").
			Optional().
			Comment("Number of nodes allocated to the cluster"),
		field.Int32("default_storage_type").
			Optional().
			Comment("Default storage type: 0=STORAGE_TYPE_UNSPECIFIED, 1=SSD, 2=HDD"),

		// JSONB fields
		field.JSON("encryption_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("cluster_config_json", json.RawMessage{}).
			Optional(),

		// Parent instance reference
		field.String("instance_name").
			NotEmpty().
			Comment("Parent instance resource name"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPBigtableCluster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
		index.Fields("instance_name"),
	}
}

func (BronzeHistoryGCPBigtableCluster) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_bigtable_clusters_history"},
	}
}
