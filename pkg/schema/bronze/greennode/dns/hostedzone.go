package dns

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeDNSHostedZone represents a GreenNode DNS hosted zone in the bronze layer.
type BronzeGreenNodeDNSHostedZone struct {
	ent.Schema
}

func (BronzeGreenNodeDNSHostedZone) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeDNSHostedZone) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Hosted zone ID"),
		field.String("domain_name").
			NotEmpty(),
		field.String("status").
			Optional(),
		field.String("description").
			Optional(),
		field.String("type").
			Optional(),
		field.Int("count_records").
			Optional(),
		field.JSON("assoc_vpc_ids_json", json.RawMessage{}).
			Optional().
			Comment("Associated VPC IDs as JSONB"),
		field.JSON("assoc_vpc_map_region_json", json.RawMessage{}).
			Optional().
			Comment("Associated VPC region mappings as JSONB"),
		field.Int("portal_user_id").
			Optional(),
		field.String("created_at_api").
			Optional().
			Comment("Hosted zone creation timestamp from API"),
		field.String("deleted_at_api").
			Optional().
			Nillable(),
		field.String("updated_at_api").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeDNSHostedZone) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("records", BronzeGreenNodeDNSRecord.Type),
	}
}

func (BronzeGreenNodeDNSHostedZone) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeDNSHostedZone) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_dns_hosted_zones"},
	}
}

// BronzeGreenNodeDNSRecord represents a DNS record within a hosted zone.
type BronzeGreenNodeDNSRecord struct {
	ent.Schema
}

func (BronzeGreenNodeDNSRecord) Fields() []ent.Field {
	return []ent.Field{
		field.String("record_id").
			NotEmpty().
			Comment("SDK record ID"),
		field.String("sub_domain").
			Optional(),
		field.String("status").
			Optional(),
		field.String("type").
			Optional(),
		field.String("routing_policy").
			Optional(),
		field.JSON("value_json", json.RawMessage{}).
			Optional().
			Comment("Record values as JSONB"),
		field.Int("ttl").
			Optional(),
		field.Bool("enable_sticky_session").
			Optional().
			Nillable(),
		field.String("created_at_api").
			Optional(),
		field.String("deleted_at_api").
			Optional().
			Nillable(),
		field.String("updated_at_api").
			Optional(),
	}
}

func (BronzeGreenNodeDNSRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("hosted_zone", BronzeGreenNodeDNSHostedZone.Type).
			Ref("records").
			Unique().
			Required(),
	}
}

func (BronzeGreenNodeDNSRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_dns_records"},
	}
}
