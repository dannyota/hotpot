package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDODomain stores historical snapshots of DigitalOcean Domains.
type BronzeHistoryDODomain struct {
	ent.Schema
}

func (BronzeHistoryDODomain) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDODomain) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Domain by resource_id"),
		field.Int("ttl").
			Default(0),
		field.String("zone_file").
			Optional(),
	}
}

func (BronzeHistoryDODomain) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryDODomain) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_domains_history"},
	}
}
