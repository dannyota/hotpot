package inventory

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryMEECInventoryInstalledSoftware stores historical snapshots of per-computer installed software.
type BronzeHistoryMEECInventoryInstalledSoftware struct {
	ent.Schema
}

func (BronzeHistoryMEECInventoryInstalledSoftware) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryMEECInventoryInstalledSoftware) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze installed software by resource_id"),

		field.String("computer_resource_id").
			NotEmpty(),
		field.Int("software_id"),
		field.String("software_name").
			Optional(),
		field.String("software_version").
			Optional(),
		field.String("display_name").
			Optional(),
		field.String("manufacturer_name").
			Optional(),
		field.Int64("installed_date").
			Optional(),
		field.String("architecture").
			Optional(),
		field.String("location").
			Optional(),
		field.Int("sw_type").
			Optional(),
		field.String("sw_category_name").
			Optional(),
		field.Int64("detected_time").
			Optional(),
	}
}

func (BronzeHistoryMEECInventoryInstalledSoftware) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("computer_resource_id"),
	}
}

func (BronzeHistoryMEECInventoryInstalledSoftware) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "meec_inventory_installed_software_history"},
	}
}
