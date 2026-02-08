package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BronzeHistoryGCPComputeNetwork stores historical snapshots of GCP VPC networks.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeNetwork struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeNetwork) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze network by resource_id"),
		field.Time("valid_from").
			Immutable().
			Comment("Start of validity period"),
		field.Time("valid_to").
			Optional().
			Nillable().
			Comment("End of validity period (null = current)"),
		field.Time("collected_at").
			Comment("Timestamp when this snapshot was collected"),

		// All network fields (same as bronze.BronzeGCPComputeNetwork)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// Network configuration
		field.Bool("auto_create_subnetworks").
			Default(false),
		field.Int("mtu").
			Default(0),
		field.String("routing_mode").
			Optional(),
		field.String("network_firewall_policy_enforcement_order").
			Optional(),

		// IPv6 configuration
		field.Bool("enable_ula_internal_ipv6").
			Default(false),
		field.String("internal_ipv6_range").
			Optional(),

		// Gateway
		field.String("gateway_ipv4").
			Optional(),

		// SubnetworksJSON
		field.JSON("subnetworks_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeNetwork) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeNetwork) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_networks_history"},
	}
}

// BronzeHistoryGCPComputeNetworkPeering stores historical snapshots of network peerings.
// Links via network_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeNetworkPeering struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeNetworkPeering) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("network_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeNetwork"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// All peering fields (same as bronze.BronzeGCPComputeNetworkPeering)
		field.String("name").
			NotEmpty(),
		field.String("network").
			Optional(),
		field.String("state").
			Optional(),
		field.String("state_details").
			Optional(),
		field.Bool("export_custom_routes").
			Default(false),
		field.Bool("import_custom_routes").
			Default(false),
		field.Bool("export_subnet_routes_with_public_ip").
			Default(false),
		field.Bool("import_subnet_routes_with_public_ip").
			Default(false),
		field.Bool("exchange_subnet_routes").
			Default(false),
		field.String("stack_type").
			Optional(),
		field.Int("peer_mtu").
			Default(0),
		field.Bool("auto_create_routes").
			Default(false),
	}
}

func (BronzeHistoryGCPComputeNetworkPeering) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("network_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeNetworkPeering) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_network_peerings_history"},
	}
}
