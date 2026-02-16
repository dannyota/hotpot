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

// BronzeHistoryGCPBigtableInstance stores historical snapshots of GCP Bigtable instances.
type BronzeHistoryGCPBigtableInstance struct {
	ent.Schema
}

func (BronzeHistoryGCPBigtableInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPBigtableInstance) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze instance by resource_id"),

		// All instance fields
		field.String("display_name").
			Optional(),
		field.Int32("state").
			Optional().
			Comment("Instance state: 0=STATE_NOT_KNOWN, 1=READY, 2=CREATING"),
		field.Int32("instance_type").
			Optional().
			Comment("Instance type: 0=TYPE_UNSPECIFIED, 1=PRODUCTION, 2=DEVELOPMENT"),
		field.String("create_time").
			Optional(),
		field.Bool("satisfies_pzs").
			Optional().
			Nillable(),

		// JSONB fields
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPBigtableInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPBigtableInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_bigtable_instances_history"},
	}
}
