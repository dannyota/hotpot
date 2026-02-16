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

// BronzeHistoryGCPAccessContextManagerServicePerimeter stores historical snapshots of service perimeters.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPAccessContextManagerServicePerimeter struct {
	ent.Schema
}

func (BronzeHistoryGCPAccessContextManagerServicePerimeter) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPAccessContextManagerServicePerimeter) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze service perimeter by resource_id"),

		// All service perimeter fields (same as bronze.BronzeGCPAccessContextManagerServicePerimeter)
		field.String("title").
			Optional(),
		field.String("description").
			Optional(),
		field.Int("perimeter_type").
			Default(0),
		field.String("etag").
			Optional(),
		field.Bool("use_explicit_dry_run_spec").
			Default(false),
		field.JSON("status_json", json.RawMessage{}).
			Optional(),
		field.JSON("spec_json", json.RawMessage{}).
			Optional(),
		field.String("access_policy_name").
			NotEmpty(),
		field.String("organization_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPAccessContextManagerServicePerimeter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPAccessContextManagerServicePerimeter) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_accesscontextmanager_service_perimeters_history"},
	}
}
