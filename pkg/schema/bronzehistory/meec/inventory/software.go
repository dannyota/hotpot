package inventory

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryMEECInventorySoftware stores historical snapshots of MEEC software catalog.
type BronzeHistoryMEECInventorySoftware struct {
	ent.Schema
}

func (BronzeHistoryMEECInventorySoftware) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryMEECInventorySoftware) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze software by resource_id"),

		field.String("software_name").
			NotEmpty(),
		field.String("software_version").
			Optional(),
		field.String("display_name").
			Optional(),
		field.Int("manufacturer_id").
			Optional(),
		field.String("manufacturer_name").
			Optional(),
		field.String("sw_category_name").
			Optional(),
		field.Int("sw_type").
			Optional(),
		field.Int("sw_family").
			Optional(),
		field.String("installed_format").
			Optional(),
		field.Int("is_usage_prohibited").
			Optional(),
		field.Int("managed_installations").
			Optional(),
		field.Int("network_installations").
			Optional(),
		field.Int("managed_sw_id").
			Optional(),
		field.Int64("detected_time").
			Optional(),
		field.String("compliant_status").
			Optional(),
		field.String("total_copies").
			Optional(),
		field.String("remaining_copies").
			Optional(),
	}
}

func (BronzeHistoryMEECInventorySoftware) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryMEECInventorySoftware) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "meec_inventory_software_history"},
	}
}
