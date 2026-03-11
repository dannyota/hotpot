package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceSoftwareMatchRule holds product matching overrides for software lifecycle detection.
// Users can edit these via Metabase to tune how installed software maps to endoflife.date products.
type BronzeReferenceSoftwareMatchRule struct {
	ent.Schema
}

func (BronzeReferenceSoftwareMatchRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceSoftwareMatchRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite key: {rule_type}:{product_slug}:{os_type}:{value}"),
		field.String("product_slug").
			NotEmpty().
			Comment("endoflife.date product slug (e.g. docker-engine, firefox)"),
		field.String("rule_type").
			NotEmpty().
			Comment("Rule type: extra_prefix, exclude, name_cycle_map"),
		field.String("os_type").
			Optional().
			Comment("OS filter: linux, windows, macos, or empty for all"),
		field.String("value").
			NotEmpty().
			Comment("The prefix, exclude word, or name-cycle key"),
		field.String("extra_value").
			Optional().
			Comment("For name_cycle_map: the cycle value (e.g. 12.0)"),
	}
}

func (BronzeReferenceSoftwareMatchRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_slug"),
		index.Fields("rule_type"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceSoftwareMatchRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_software_match_rules"},
	}
}
