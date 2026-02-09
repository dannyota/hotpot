package compute

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

// BronzeGCPComputeNetwork represents a GCP VPC network in the bronze layer.
// Fields preserve raw API response data from compute.networks.list.
type BronzeGCPComputeNetwork struct {
	ent.Schema
}

func (BronzeGCPComputeNetwork) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeNetwork) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID, used as primary key for linking"),
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

		// SubnetworksJSON contains list of subnetwork URLs in this network.
		//
		//	["projects/.../regions/.../subnetworks/subnet1", ...]
		field.JSON("subnetworks_json", json.RawMessage{}).
			Optional().
			Comment("List of subnetwork URLs in this network"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeNetwork) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("peerings", BronzeGCPComputeNetworkPeering.Type),
	}
}

func (BronzeGCPComputeNetwork) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeNetwork) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_networks"},
	}
}

// BronzeGCPComputeNetworkPeering represents a VPC network peering connection.
// Data from network.peerings[].
type BronzeGCPComputeNetworkPeering struct {
	ent.Schema
}

func (BronzeGCPComputeNetworkPeering) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields
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

func (BronzeGCPComputeNetworkPeering) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("network_ref", BronzeGCPComputeNetwork.Type).
			Ref("peerings").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeNetworkPeering) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_network_peerings"},
	}
}
