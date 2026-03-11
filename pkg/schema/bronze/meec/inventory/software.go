package inventory

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeMEECInventorySoftware represents a global software entry in the MEEC inventory catalog.
type BronzeMEECInventorySoftware struct {
	ent.Schema
}

func (BronzeMEECInventorySoftware) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeMEECInventorySoftware) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("MEEC software_id (integer as string)"),
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
			Optional().
			Comment("0=Unidentified, 1=Commercial, 2=Non-commercial"),
		field.Int("sw_family").
			Optional(),
		field.String("installed_format").
			Optional().
			Comment("EXE, MSI, etc."),
		field.Int("is_usage_prohibited").
			Optional().
			Comment("0=Not assigned, 1=Allowed, 2=Prohibited"),
		field.Int("managed_installations").
			Optional(),
		field.Int("network_installations").
			Optional(),
		field.Int("managed_sw_id").
			Optional(),
		field.Int64("detected_time").
			Optional().
			Comment("Timestamp in milliseconds"),
		field.String("compliant_status").
			Optional(),
		field.String("total_copies").
			Optional(),
		field.String("remaining_copies").
			Optional(),
	}
}

func (BronzeMEECInventorySoftware) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("software_name"),
		index.Fields("manufacturer_name"),
		index.Fields("collected_at"),
	}
}

func (BronzeMEECInventorySoftware) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "meec_inventory_software"},
	}
}
