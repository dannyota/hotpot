package filestore

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPFilestoreInstance stores historical snapshots of GCP Filestore instances.
type BronzeHistoryGCPFilestoreInstance struct {
	ent.Schema
}

func (BronzeHistoryGCPFilestoreInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPFilestoreInstance) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze instance by resource_id"),

		// All instance fields
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Int("state").
			Default(0).
			Comment("Instance state (0=UNSPECIFIED, 1=CREATING, 2=READY, 3=REPAIRING, 4=DELETING, 6=ERROR, 7=RESTORING, 8=SUSPENDED, 9=SUSPENDING, 10=RESUMING, 11=REVERTING, 12=PROMOTING)"),
		field.String("status_message").
			Optional(),
		field.String("create_time").
			Optional(),
		field.Int("tier").
			Default(0).
			Comment("Instance tier (0=UNSPECIFIED, 1=STANDARD, 2=PREMIUM, 3=BASIC_HDD, 4=BASIC_SSD, 5=HIGH_SCALE_SSD, 6=ENTERPRISE, 7=ZONAL, 8=REGIONAL)"),

		// JSONB fields
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),
		field.JSON("file_shares_json", json.RawMessage{}).
			Optional(),
		field.JSON("networks_json", json.RawMessage{}).
			Optional(),

		field.String("etag").
			Optional(),
		field.Bool("satisfies_pzs").
			Default(false),
		field.Bool("satisfies_pzi").
			Default(false),
		field.String("kms_key_name").
			Optional(),

		field.JSON("suspension_reasons_json", json.RawMessage{}).
			Optional(),

		field.Int64("max_capacity_gb").
			Optional(),
		field.Int("protocol").
			Default(0).
			Comment("File access protocol (0=UNSPECIFIED, 1=NFS_V3, 2=NFS_V4_1)"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			Optional(),
	}
}

func (BronzeHistoryGCPFilestoreInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPFilestoreInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_filestore_instances_history"},
	}
}
