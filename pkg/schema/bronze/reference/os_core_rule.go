package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceOSCoreRule holds OS core classification patterns for software lifecycle detection.
// Packages matching these rules have their lifecycle tied to the OS release, not tracked independently.
// Users can edit these via Metabase to tune classification.
type BronzeReferenceOSCoreRule struct {
	ent.Schema
}

func (BronzeReferenceOSCoreRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceOSCoreRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite key: {rule_type}:{os_type}:{value}"),
		field.String("rule_type").
			NotEmpty().
			Comment("Rule type: prefix, suffix, exact"),
		field.String("os_type").
			Optional().
			Comment("OS filter: linux, windows, macos, or empty for all"),
		field.String("value").
			NotEmpty().
			Comment("The pattern (e.g. linux-, -keyring, gmail)"),
		field.String("description").
			Optional().
			Comment("Human-readable explanation of why this is OS core"),
	}
}

func (BronzeReferenceOSCoreRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("rule_type"),
		index.Fields("os_type"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceOSCoreRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_os_core_rules"},
	}
}
