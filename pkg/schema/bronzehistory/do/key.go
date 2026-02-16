package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDOKey stores historical snapshots of DigitalOcean SSH Keys.
type BronzeHistoryDOKey struct {
	ent.Schema
}

func (BronzeHistoryDOKey) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDOKey) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Key by resource_id"),
		field.String("name").
			NotEmpty(),
		field.String("fingerprint").
			Optional(),
		field.String("public_key").
			Optional(),
	}
}

func (BronzeHistoryDOKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("fingerprint"),
	}
}

func (BronzeHistoryDOKey) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_keys_history"},
	}
}
