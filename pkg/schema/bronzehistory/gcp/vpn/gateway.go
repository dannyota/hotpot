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

// BronzeHistoryGCPVPNGateway stores historical snapshots of GCP Compute Engine VPN gateways.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPVPNGateway struct {
	ent.Schema
}

func (BronzeHistoryGCPVPNGateway) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPVPNGateway) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze VPN gateway by resource_id"),

		// All VPN gateway fields (same as bronze.BronzeGCPVPNGateway)
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

		// JSONB fields
		field.JSON("vpn_interfaces_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPVPNGateway) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPVPNGateway) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_vpn_gateways_history"},
	}
}

// BronzeHistoryGCPVPNGatewayLabel stores historical snapshots of VPN gateway labels.
// Links via vpn_gateway_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPVPNGatewayLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPVPNGatewayLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("vpn_gateway_history_id").
			Comment("Links to parent BronzeHistoryGCPVPNGateway"),
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

func (BronzeHistoryGCPVPNGatewayLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("vpn_gateway_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPVPNGatewayLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_vpn_gateway_labels_history"},
	}
}
