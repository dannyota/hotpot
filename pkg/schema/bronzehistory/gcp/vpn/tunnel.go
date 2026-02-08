package vpn

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPVPNTunnel stores historical versions of VPN tunnels.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPVPNTunnel struct {
	ent.Schema
}

func (BronzeHistoryGCPVPNTunnel) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPVPNTunnel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze VPN tunnel by resource_id"),

		// All VPN tunnel fields (same as bronze.BronzeGCPVPNTunnel)
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

		// Security
		field.String("shared_secret_hash").
			Optional(),

		// Gateway references
		field.String("vpn_gateway").
			Optional(),
		field.String("target_vpn_gateway").
			Optional(),
		field.Int32("vpn_gateway_interface").
			Optional(),

		// JSONB fields
		field.JSON("local_traffic_selector_json", json.RawMessage{}).
			Optional(),
		field.JSON("remote_traffic_selector_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPVPNTunnel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPVPNTunnel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_vpn_tunnels_history"},
	}
}

// BronzeHistoryGCPVPNTunnelLabel stores historical versions of VPN tunnel labels.
// Links via vpn_tunnel_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPVPNTunnelLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPVPNTunnelLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("vpn_tunnel_history_id").
			Comment("Links to parent BronzeHistoryGCPVPNTunnel"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Label fields
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPVPNTunnelLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("vpn_tunnel_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPVPNTunnelLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_vpn_tunnel_labels_history"},
	}
}
