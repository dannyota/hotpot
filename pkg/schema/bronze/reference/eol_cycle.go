package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceEOLCycle represents a release cycle from the endoflife.date database.
type BronzeReferenceEOLCycle struct {
	ent.Schema
}

func (BronzeReferenceEOLCycle) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceEOLCycle) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite key: {product}:{cycle} (e.g. rhel:9)"),
		field.String("product").
			Comment("Product slug (e.g. rhel)"),
		field.String("cycle").
			Comment("Release cycle identifier (e.g. 9, 24.04, 2022)"),
		field.Time("release_date").
			Optional().
			Comment("General availability date"),
		field.Time("eoas").
			Optional().
			Comment("End of Active Support date"),
		field.Time("eol").
			Optional().
			Comment("End of Life date"),
		field.Time("eoes").
			Optional().
			Comment("End of Extended Support date"),
		field.String("latest").
			Optional().
			Comment("Latest version string"),
		field.Time("latest_release_date").
			Optional().
			Comment("Date of latest release"),
		field.Time("lts").
			Optional().
			Comment("LTS end date (nil if not LTS or no date available)"),
	}
}

func (BronzeReferenceEOLCycle) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceEOLCycle) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_eol_cycles"},
	}
}
