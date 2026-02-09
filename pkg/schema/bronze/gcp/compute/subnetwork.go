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

// BronzeGCPComputeSubnetwork represents a GCP VPC subnetwork in the bronze layer.
// Fields preserve raw API response data from compute.subnetworks.aggregatedList.
type BronzeGCPComputeSubnetwork struct {
	ent.Schema
}

func (BronzeGCPComputeSubnetwork) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeSubnetwork) Fields() []ent.Field {
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

		// Network relationship (URL to parent network)
		field.String("network").
			NotEmpty(),
		field.String("region").
			NotEmpty(),

		// IP configuration
		field.String("ip_cidr_range").
			NotEmpty(),
		field.String("gateway_address").
			Optional(),

		// Purpose and role
		field.String("purpose").
			Optional(),
		field.String("role").
			Optional(),

		// Private Google Access
		field.Bool("private_ip_google_access").
			Default(false),
		field.String("private_ipv6_google_access").
			Optional(),

		// Stack type and IPv6
		field.String("stack_type").
			Optional(),
		field.String("ipv6_access_type").
			Optional(),
		field.String("internal_ipv6_prefix").
			Optional(),
		field.String("external_ipv6_prefix").
			Optional(),

		// LogConfigJSON contains VPC flow logging configuration.
		//
		//	{
		//	  "enable": true,
		//	  "aggregationInterval": "INTERVAL_5_SEC",
		//	  "flowSampling": 0.5,
		//	  "metadata": "INCLUDE_ALL_METADATA"
		//	}
		field.JSON("log_config_json", json.RawMessage{}).
			Optional().
			Comment("VPC flow logging configuration"),

		// Fingerprint for optimistic locking
		field.String("fingerprint").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeSubnetwork) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("secondary_ip_ranges", BronzeGCPComputeSubnetworkSecondaryRange.Type),
	}
}

func (BronzeGCPComputeSubnetwork) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeSubnetwork) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_subnetworks"},
	}
}

// BronzeGCPComputeSubnetworkSecondaryRange represents a secondary IP range on a subnetwork.
// Data from subnetwork.secondaryIpRanges[].
type BronzeGCPComputeSubnetworkSecondaryRange struct {
	ent.Schema
}

func (BronzeGCPComputeSubnetworkSecondaryRange) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields
		field.String("range_name").
			NotEmpty(),
		field.String("ip_cidr_range").
			NotEmpty(),
	}
}

func (BronzeGCPComputeSubnetworkSecondaryRange) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("subnetwork", BronzeGCPComputeSubnetwork.Type).
			Ref("secondary_ip_ranges").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeSubnetworkSecondaryRange) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_subnetwork_secondary_ranges"},
	}
}
