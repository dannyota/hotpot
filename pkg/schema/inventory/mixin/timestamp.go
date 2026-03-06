package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Timestamp provides collection metadata for inventory layer.
type Timestamp struct {
	mixin.Schema
}

func (Timestamp) Fields() []ent.Field {
	return []ent.Field{
		field.Time("collected_at"),
		field.Time("first_collected_at").Immutable(),
		field.Time("normalized_at"),
	}
}
