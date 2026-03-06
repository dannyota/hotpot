package lifecycle

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	goldmixin "danny.vn/hotpot/pkg/schema/gold/mixin"
)

// GoldLifecycleSoftware holds per-machine software EOL classification and status.
// Each row represents one (machine_id, name) pair from inventory.software,
// classified as matched (with EOL data), os_core, or unmatched.
type GoldLifecycleSoftware struct {
	ent.Schema
}

func (GoldLifecycleSoftware) Mixin() []ent.Mixin {
	return []ent.Mixin{
		goldmixin.Timestamp{},
	}
}

func (GoldLifecycleSoftware) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("machine_id").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("version").Optional(),

		// Classification: matched, os_core, unmatched.
		field.String("classification").NotEmpty(),

		// EOL product info (populated for classification=matched).
		field.String("eol_product_slug").Optional(),
		field.String("eol_product_name").Optional(),
		field.String("eol_category").Optional(),
		field.String("eol_cycle").Optional(),
		field.Time("eol_date").Optional(),
		field.Time("eoas_date").Optional(),
		field.Time("eoes_date").Optional(),

		// EOL status: active, eoas_expired, eol_expired, eoes_expired, unknown.
		field.String("eol_status").NotEmpty(),

		field.String("latest_version").Optional(),
	}
}

func (GoldLifecycleSoftware) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("machine_id"),
		index.Fields("classification"),
		index.Fields("eol_status"),
		index.Fields("eol_product_slug"),
		index.Fields("machine_id", "name").Unique(),
	}
}

func (GoldLifecycleSoftware) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "lifecycle_software"},
	}
}
