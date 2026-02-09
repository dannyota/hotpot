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

// BronzeGCPVPNTargetGateway represents a GCP Compute Engine Classic VPN gateway in the bronze layer.
// Fields preserve raw API response data from compute.targetVpnGateways.aggregatedList.
type BronzeGCPVPNTargetGateway struct {
	ent.Schema
}

func (BronzeGCPVPNTargetGateway) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPVPNTargetGateway) Fields() []ent.Field {
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
		field.String("status").
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

		// ForwardingRulesJSON contains forwarding rule URLs.
		//
		//	["https://www.googleapis.com/compute/v1/projects/.../forwardingRules/..."]
		field.JSON("forwarding_rules_json", json.RawMessage{}).
			Optional(),

		// TunnelsJSON contains VPN tunnel URLs.
		//
		//	["https://www.googleapis.com/compute/v1/projects/.../vpnTunnels/..."]
		field.JSON("tunnels_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPVPNTargetGateway) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPVPNTargetGatewayLabel.Type),
	}
}

func (BronzeGCPVPNTargetGateway) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("status"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPVPNTargetGateway) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_vpn_gateways"},
	}
}

// BronzeGCPVPNTargetGatewayLabel represents a label attached to a Classic VPN gateway.
type BronzeGCPVPNTargetGatewayLabel struct {
	ent.Schema
}

func (BronzeGCPVPNTargetGatewayLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPVPNTargetGatewayLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("target_vpn_gateway", BronzeGCPVPNTargetGateway.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPVPNTargetGatewayLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_vpn_gateway_labels"},
	}
}
