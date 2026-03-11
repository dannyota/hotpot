package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Timestamp provides temporal fields for parent bronze history schemas.
type Timestamp struct {
	mixin.Schema
}

func (Timestamp) Fields() []ent.Field {
	return []ent.Field{
		field.Time("valid_from").
			Immutable().
			Comment("Start of validity period"),
		field.Time("valid_to").
			Optional().
			Nillable().
			Comment("End of validity period (null = current)"),
		field.Time("collected_at").
			Comment("Timestamp when this snapshot was collected"),
		field.Time("first_collected_at").
			Comment("Timestamp when this asset was first collected"),
	}
}
