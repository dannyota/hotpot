package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDODomainRecord stores historical snapshots of DigitalOcean Domain Records.
type BronzeHistoryDODomainRecord struct {
	ent.Schema
}

func (BronzeHistoryDODomainRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDODomainRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze DomainRecord by resource_id"),
		field.String("domain_name").
			NotEmpty(),
		field.Int("record_id"),
		field.String("type").
			Optional(),
		field.String("name").
			Optional(),
		field.String("data").
			Optional(),
		field.Int("priority").
			Default(0),
		field.Int("port").
			Default(0),
		field.Int("ttl").
			Default(0),
		field.Int("weight").
			Default(0),
		field.Int("flags").
			Default(0),
		field.String("tag").
			Optional(),
	}
}

func (BronzeHistoryDODomainRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("domain_name"),
		index.Fields("type"),
	}
}

func (BronzeHistoryDODomainRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_domain_records_history"},
	}
}
