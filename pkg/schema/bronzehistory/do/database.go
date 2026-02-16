package do

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDODatabase stores historical snapshots of DigitalOcean Database Clusters.
type BronzeHistoryDODatabase struct {
	ent.Schema
}

func (BronzeHistoryDODatabase) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDODatabase) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Database by resource_id"),
		field.String("name").
			Optional(),
		field.String("engine_slug").
			Optional(),
		field.String("version_slug").
			Optional(),
		field.Int("num_nodes").
			Default(0),
		field.String("size_slug").
			Optional(),
		field.String("region_slug").
			Optional(),
		field.String("status").
			Optional(),
		field.String("project_id").
			Optional(),
		field.Uint64("storage_size_mib").
			Default(0),
		field.String("private_network_uuid").
			Optional(),
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),
		field.JSON("maintenance_window_json", json.RawMessage{}).
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
	}
}

func (BronzeHistoryDODatabase) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("engine_slug"),
		index.Fields("region_slug"),
		index.Fields("status"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryDODatabase) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_databases_history"},
	}
}
