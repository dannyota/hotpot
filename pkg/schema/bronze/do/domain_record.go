package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDODomainRecord represents a DigitalOcean Domain Record in the bronze layer.
type BronzeDODomainRecord struct {
	ent.Schema
}

func (BronzeDODomainRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDODomainRecord) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite ID: {domain}:{recordID}"),
		field.String("domain_name").
			NotEmpty(),
		field.Int("record_id").
			Comment("DigitalOcean record ID"),
		field.String("type").
			Optional().
			Comment("DNS record type (A, AAAA, CNAME, etc.)"),
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

func (BronzeDODomainRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("domain_name"),
		index.Fields("type"),
		index.Fields("collected_at"),
	}
}

func (BronzeDODomainRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_domain_records"},
	}
}
