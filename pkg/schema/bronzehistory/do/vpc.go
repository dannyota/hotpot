package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDOVpc stores historical snapshots of DigitalOcean VPCs.
type BronzeHistoryDOVpc struct {
	ent.Schema
}

func (BronzeHistoryDOVpc) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDOVpc) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze VPC by resource_id"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("region").
			Optional(),
		field.String("ip_range").
			Optional(),
		field.String("urn").
			Optional(),
		field.Bool("is_default").
			Default(false),
		field.Time("api_created_at").
			Optional().
			Nillable(),
	}
}

func (BronzeHistoryDOVpc) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("region"),
	}
}

func (BronzeHistoryDOVpc) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_vpcs_history"},
	}
}
