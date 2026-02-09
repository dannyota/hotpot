package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDOVpc represents a DigitalOcean VPC in the bronze layer.
type BronzeDOVpc struct {
	ent.Schema
}

func (BronzeDOVpc) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDOVpc) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("DigitalOcean VPC UUID"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("region").
			Optional().
			Comment("Region slug (e.g. nyc1)"),
		field.String("ip_range").
			Optional().
			Comment("CIDR block"),
		field.String("urn").
			Optional(),
		field.Bool("is_default").
			Default(false),
		field.Time("api_created_at").
			Optional().
			Nillable(),
	}
}

func (BronzeDOVpc) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("region"),
		index.Fields("is_default"),
		index.Fields("collected_at"),
	}
}

func (BronzeDOVpc) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_vpcs"},
	}
}
