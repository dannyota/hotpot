package dns

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPDNSPolicy represents a GCP Cloud DNS policy in the bronze layer.
// Fields preserve raw API response data from dns.policies.list.
type BronzeGCPDNSPolicy struct {
	ent.Schema
}

func (BronzeGCPDNSPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPDNSPolicy) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields - ID is uint64, stored as string
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID (uint64 as string), used as primary key for linking"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Bool("enable_inbound_forwarding").
			Default(false),
		field.Bool("enable_logging").
			Default(false),

		// JSONB configuration fields
		// NetworksJSON contains the list of networks this policy is applied to.
		//
		//	[{"networkUrl": "https://www.googleapis.com/compute/v1/projects/.../networks/..."}]
		field.JSON("networks_json", json.RawMessage{}).
			Optional(),

		// AlternativeNameServerConfigJSON contains alternative name server configuration.
		//
		//	{"targetNameServers": [{"ipv4Address": "8.8.8.8", "forwardingPath": "default"}]}
		field.JSON("alternative_name_server_config_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPDNSPolicy) Edges() []ent.Edge {
	return nil
}

func (BronzeGCPDNSPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPDNSPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_dns_policies"},
	}
}
