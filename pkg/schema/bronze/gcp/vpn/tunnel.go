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

// BronzeGCPVPNTunnel represents a GCP Compute Engine VPN tunnel in the bronze layer.
// Fields preserve raw API response data from compute.vpnTunnels.aggregatedList.
type BronzeGCPVPNTunnel struct {
	ent.Schema
}

func (BronzeGCPVPNTunnel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPVPNTunnel) Fields() []ent.Field {
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
		field.String("detailed_status").
			Optional(),
		field.String("region").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("label_fingerprint").
			Optional(),

		// IKE settings
		field.Int32("ike_version").
			Optional(),

		// Peer settings
		field.String("peer_ip").
			Optional(),
		field.String("peer_external_gateway").
			Optional(),
		field.Int32("peer_external_gateway_interface").
			Optional(),
		field.String("peer_gcp_gateway").
			Optional(),

		// Routing
		field.String("router").
			Optional(),

		// Security (SharedSecret excluded - sensitive)
		field.String("shared_secret_hash").
			Optional(),

		// Gateway references
		field.String("vpn_gateway").
			Optional(),
		field.String("target_vpn_gateway").
			Optional(),
		field.Int32("vpn_gateway_interface").
			Optional(),

		// LocalTrafficSelectorJSON contains local CIDR ranges.
		//
		//	["10.0.0.0/8", "192.168.0.0/16"]
		field.JSON("local_traffic_selector_json", json.RawMessage{}).
			Optional(),

		// RemoteTrafficSelectorJSON contains remote CIDR ranges.
		//
		//	["172.16.0.0/12"]
		field.JSON("remote_traffic_selector_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPVPNTunnel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPVPNTunnelLabel.Type),
	}
}

func (BronzeGCPVPNTunnel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("status"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPVPNTunnel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_vpn_tunnels"},
	}
}

// BronzeGCPVPNTunnelLabel represents a label attached to a VPN tunnel.
type BronzeGCPVPNTunnelLabel struct {
	ent.Schema
}

func (BronzeGCPVPNTunnelLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPVPNTunnelLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("vpn_tunnel", BronzeGCPVPNTunnel.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPVPNTunnelLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_vpn_tunnel_labels"},
	}
}
