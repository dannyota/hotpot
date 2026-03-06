package inventory

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeMEECInventoryInstalledSoftware represents software installed on a specific computer.
type BronzeMEECInventoryInstalledSoftware struct {
	ent.Schema
}

func (BronzeMEECInventoryInstalledSoftware) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeMEECInventoryInstalledSoftware) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Synthesized: computerResourceID||softwareID"),
		field.String("computer_resource_id").
			NotEmpty().
			Comment("MEEC computer resource_id (denormalized)"),
		field.Int("software_id").
			Comment("MEEC software_id"),
		field.String("software_name").
			Optional(),
		field.String("software_version").
			Optional(),
		field.String("display_name").
			Optional(),
		field.String("manufacturer_name").
			Optional(),
		field.Int64("installed_date").
			Optional().
			Comment("Timestamp in milliseconds"),
		field.String("architecture").
			Optional().
			Comment("32-bit or 64-bit"),
		field.String("location").
			Optional(),
		field.Int("sw_type").
			Optional().
			Comment("0=Unidentified, 1=Commercial, 2=Non-commercial"),
		field.String("sw_category_name").
			Optional(),
		field.Int64("detected_time").
			Optional().
			Comment("Timestamp in milliseconds"),
	}
}

func (BronzeMEECInventoryInstalledSoftware) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("computer_resource_id"),
		index.Fields("software_name"),
		index.Fields("collected_at"),
	}
}

func (BronzeMEECInventoryInstalledSoftware) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "meec_inventory_installed_software"},
	}
}
