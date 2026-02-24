package dns

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeDNSHostedZone stores historical snapshots of GreenNode DNS hosted zones.
type BronzeHistoryGreenNodeDNSHostedZone struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeDNSHostedZone) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeDNSHostedZone) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
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
			Optional(),
		field.JSON("assoc_vpc_map_region_json", json.RawMessage{}).
			Optional(),
		field.Int("portal_user_id").
			Optional(),
		field.String("created_at_api").
			Optional(),
		field.String("deleted_at_api").
			Optional().
			Nillable(),
		field.String("updated_at_api").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeDNSHostedZone) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeDNSHostedZone) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_dns_hosted_zones_history"},
	}
}

// BronzeHistoryGreenNodeDNSRecord stores historical snapshots of DNS records.
type BronzeHistoryGreenNodeDNSRecord struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeDNSRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.Uint("hosted_zone_history_id").
			Comment("Links to parent BronzeHistoryGreenNodeDNSHostedZone"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),
		field.String("record_id").
			NotEmpty(),
		field.String("sub_domain").
			Optional(),
		field.String("status").
			Optional(),
		field.String("type").
			Optional(),
		field.String("routing_policy").
			Optional(),
		field.JSON("value_json", json.RawMessage{}).
			Optional(),
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

func (BronzeHistoryGreenNodeDNSRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("hosted_zone_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGreenNodeDNSRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_dns_records_history"},
	}
}
