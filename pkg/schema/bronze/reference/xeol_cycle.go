package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceXeolCycle represents a release cycle from the xeol EOL database.
type BronzeReferenceXeolCycle struct {
	ent.Schema
}

func (BronzeReferenceXeolCycle) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceXeolCycle) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite key: {product_id}:{sqlite_cycle_id}"),
		field.String("product_id").
			Comment("Reference to xeol product ID"),
		field.String("release_cycle").
			Comment("Release cycle identifier"),
		field.Time("eol").
			Optional().
			Comment("End of Life date"),
		field.Bool("eol_bool").
			Default(false).
			Comment("True when EOL status is boolean (no date available)"),
		field.String("latest_release").
			Optional().
			Comment("Latest release version in this cycle"),
		field.Time("latest_release_date").
			Optional().
			Comment("Date of latest release"),
		field.Time("release_date").
			Optional().
			Comment("Release date of this cycle"),
	}
}

func (BronzeReferenceXeolCycle) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceXeolCycle) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_xeol_cycles"},
	}
}
