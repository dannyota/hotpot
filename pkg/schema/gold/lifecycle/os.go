package lifecycle

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	goldmixin "danny.vn/hotpot/pkg/schema/gold/mixin"
)

// GoldLifecycleOS holds per-machine OS EOL classification and status.
// Each row represents one machine from inventory.machines, matched against
// endoflife.date OS products (or unmatched with eol_status=unknown).
type GoldLifecycleOS struct {
	ent.Schema
}

func (GoldLifecycleOS) Mixin() []ent.Mixin {
	return []ent.Mixin{
		goldmixin.Timestamp{},
	}
}

func (GoldLifecycleOS) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("machine_id").NotEmpty(),
		field.String("hostname").Optional(),
		field.String("os_type").Optional(),
		field.String("os_name").Optional(),

		// EOL product info (populated when matched).
		field.String("eol_product_slug").Optional(),
		field.String("eol_product_name").Optional(),
		field.String("eol_cycle").Optional(),
		field.Time("eol_date").Optional(),
		field.Time("eoas_date").Optional(),
		field.Time("eoes_date").Optional(),

		// EOL status: active, eoas_expired, eol_expired, eoes_expired, unknown.
		field.String("eol_status").NotEmpty(),

		field.String("latest_version").Optional(),
	}
}

func (GoldLifecycleOS) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("machine_id").Unique(),
		index.Fields("eol_status"),
		index.Fields("eol_product_slug"),
	}
}

func (GoldLifecycleOS) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "lifecycle_os"},
	}
}
