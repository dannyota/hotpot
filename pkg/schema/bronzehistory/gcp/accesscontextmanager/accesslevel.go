package accesscontextmanager

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPAccessContextManagerAccessLevel stores historical snapshots of access levels.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPAccessContextManagerAccessLevel struct {
	ent.Schema
}

func (BronzeHistoryGCPAccessContextManagerAccessLevel) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPAccessContextManagerAccessLevel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze access level by resource_id"),

		// All access level fields (same as bronze.BronzeGCPAccessContextManagerAccessLevel)
		field.String("title").
			Optional(),
		field.String("description").
			Optional(),
		field.JSON("basic_json", json.RawMessage{}).
			Optional(),
		field.JSON("custom_json", json.RawMessage{}).
			Optional(),
		field.String("access_policy_name").
			NotEmpty(),
		field.String("organization_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPAccessContextManagerAccessLevel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPAccessContextManagerAccessLevel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_accesscontextmanager_access_levels_history"},
	}
}
