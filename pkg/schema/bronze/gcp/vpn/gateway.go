package vpn

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPVPNGateway represents a GCP Compute Engine VPN gateway (HA) in the bronze layer.
// Fields preserve raw API response data from compute.vpnGateways.aggregatedList.
type BronzeGCPVPNGateway struct {
	ent.Schema
}

func (BronzeGCPVPNGateway) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPVPNGateway) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID, used as primary key for linking"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("region").
			Optional(),
		field.String("network").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("label_fingerprint").
			Optional(),
		field.String("gateway_ip_version").
			Optional(),
		field.String("stack_type").
			Optional(),

		// VpnInterfacesJSON contains the VPN gateway interfaces configuration.
		//
		//	[{"id": 0, "ipAddress": "1.2.3.4"}, {"id": 1, "ipAddress": "5.6.7.8"}]
		field.JSON("vpn_interfaces_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPVPNGateway) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPVPNGatewayLabel.Type),
	}
}

func (BronzeGCPVPNGateway) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPVPNGateway) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_vpn_gateways"},
	}
}

// BronzeGCPVPNGatewayLabel represents a label attached to a GCP Compute VPN gateway.
type BronzeGCPVPNGatewayLabel struct {
	ent.Schema
}

func (BronzeGCPVPNGatewayLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPVPNGatewayLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("vpn_gateway", BronzeGCPVPNGateway.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPVPNGatewayLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_vpn_gateway_labels"},
	}
}
