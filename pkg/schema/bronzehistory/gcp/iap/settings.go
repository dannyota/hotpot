package iap

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPIAPSettings stores historical snapshots of IAP settings.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPIAPSettings struct {
	ent.Schema
}

func (BronzeHistoryGCPIAPSettings) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPIAPSettings) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze IAP settings by resource_id"),

		// All settings fields (same as bronze.BronzeGCPIAPSettings)
		field.String("name").
			NotEmpty(),
		field.JSON("access_settings_json", json.RawMessage{}).
			Optional(),
		field.JSON("application_settings_json", json.RawMessage{}).
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPIAPSettings) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPIAPSettings) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_iap_settings_history"},
	}
}
