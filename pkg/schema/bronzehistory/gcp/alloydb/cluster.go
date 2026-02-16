package alloydb

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPAlloyDBCluster stores historical snapshots of GCP AlloyDB clusters.
type BronzeHistoryGCPAlloyDBCluster struct {
	ent.Schema
}

func (BronzeHistoryGCPAlloyDBCluster) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPAlloyDBCluster) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze cluster by resource_id"),

		// All cluster fields
		field.String("name").
			NotEmpty(),
		field.String("display_name").
			Optional(),
		field.String("uid").
			Optional().
			Comment("System-generated UID"),
		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),
		field.String("delete_time").
			Optional(),

		// JSONB fields
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		field.Int("state").
			Optional().
			Comment("Cluster state (0=UNSPECIFIED, 1=READY, 2=STOPPED, 3=EMPTY, 4=CREATING, 5=DELETING, 6=FAILED, 7=BOOTSTRAPPING, 8=MAINTENANCE, 9=PROMOTING)"),

		field.Int("cluster_type").
			Optional().
			Comment("Cluster type (0=UNSPECIFIED, 1=PRIMARY, 2=SECONDARY)"),

		field.Int("database_version").
			Optional().
			Comment("Database engine major version (0=UNSPECIFIED, 1=POSTGRES_13, 2=POSTGRES_14, 3=POSTGRES_15, 4=POSTGRES_16)"),

		field.JSON("network_config_json", json.RawMessage{}).
			Optional(),

		field.String("network").
			Optional().
			Comment("Deprecated: VPC network resource link"),

		field.String("etag").
			Optional(),

		field.JSON("annotations_json", json.RawMessage{}).
			Optional(),

		field.Bool("reconciling").
			Optional().
			Comment("Whether the cluster is being reconciled"),

		field.JSON("initial_user_json", json.RawMessage{}).
			Optional(),
		field.JSON("automated_backup_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("ssl_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("encryption_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("encryption_info_json", json.RawMessage{}).
			Optional(),
		field.JSON("continuous_backup_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("continuous_backup_info_json", json.RawMessage{}).
			Optional(),
		field.JSON("secondary_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("primary_config_json", json.RawMessage{}).
			Optional(),

		field.Bool("satisfies_pzs").
			Optional().
			Comment("Reserved for future use"),

		field.JSON("psc_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("maintenance_update_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("maintenance_schedule_json", json.RawMessage{}).
			Optional(),

		field.Int("subscription_type").
			Optional().
			Comment("Subscription type (0=UNSPECIFIED, 1=STANDARD, 2=TRIAL)"),

		field.JSON("trial_metadata_json", json.RawMessage{}).
			Optional(),
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPAlloyDBCluster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPAlloyDBCluster) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_alloydb_clusters_history"},
	}
}
