package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Timestamp provides detection metadata for gold layer.
type Timestamp struct {
	mixin.Schema
}

func (Timestamp) Fields() []ent.Field {
	return []ent.Field{
		field.Time("detected_at"),
		field.Time("first_detected_at").Immutable(),
	}
}
