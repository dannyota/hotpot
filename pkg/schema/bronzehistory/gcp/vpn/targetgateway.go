package vpn

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BronzeHistoryGCPVPNTargetGateway stores historical versions of Classic VPN gateways.
type BronzeHistoryGCPVPNTargetGateway struct {
	ent.Schema
}

func (BronzeHistoryGCPVPNTargetGateway) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze target VPN gateway by resource_id"),
		field.Time("valid_from").
			Immutable().
			Comment("Start of validity period"),
		field.Time("valid_to").
			Optional().
			Nillable().
			Comment("End of validity period (null = current)"),
		field.Time("collected_at").
			Comment("Timestamp when this snapshot was collected"),

		// All target VPN gateway fields (same as bronze.BronzeGCPVPNTargetGateway)
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

		// JSONB fields
		field.JSON("forwarding_rules_json", json.RawMessage{}).
			Optional(),
		field.JSON("tunnels_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPVPNTargetGateway) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPVPNTargetGateway) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_target_vpn_gateways_history"},
	}
}

// BronzeHistoryGCPVPNTargetGatewayLabel stores historical versions of Classic VPN gateway labels.
type BronzeHistoryGCPVPNTargetGatewayLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPVPNTargetGatewayLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("target_vpn_gateway_history_id").
			Comment("Links to parent BronzeHistoryGCPVPNTargetGateway"),
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

func (BronzeHistoryGCPVPNTargetGatewayLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("target_vpn_gateway_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPVPNTargetGatewayLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_target_vpn_gateway_labels_history"},
	}
}
