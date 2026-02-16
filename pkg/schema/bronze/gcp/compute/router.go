package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeRouter represents a GCP Compute Engine router in the bronze layer.
// Fields preserve raw API response data from compute.routers.aggregatedList.
type BronzeGCPComputeRouter struct {
	ent.Schema
}

func (BronzeGCPComputeRouter) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeRouter) Fields() []ent.Field {
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

		// Router configuration
		field.String("network").
			Optional().
			Comment("Network URL this router belongs to"),
		field.String("region").
			Optional().
			Comment("Region URL"),

		// BGP configuration (extracted from router.bgp)
		field.Int("bgp_asn").
			Default(0).
			Comment("BGP autonomous system number"),
		field.String("bgp_advertise_mode").
			Optional().
			Comment("DEFAULT or CUSTOM"),

		// BgpAdvertisedGroupsJSON contains BGP advertised groups.
		//
		//	["ALL_SUBNETS"]
		field.JSON("bgp_advertised_groups_json", json.RawMessage{}).
			Optional(),

		// BgpAdvertisedIpRangesJSON contains custom BGP advertised IP ranges.
		//
		//	[{"range": "10.0.0.0/8", "description": "..."}]
		field.JSON("bgp_advertised_ip_ranges_json", json.RawMessage{}).
			Optional(),

		field.Int("bgp_keepalive_interval").
			Default(0).
			Comment("BGP keepalive interval in seconds"),

		// BgpPeersJSON contains BGP peer configurations.
		//
		//	[{"name": "peer1", "peerAsn": 65001, "ipAddress": "..."}]
		field.JSON("bgp_peers_json", json.RawMessage{}).
			Optional(),

		// InterfacesJSON contains router interface configurations.
		//
		//	[{"name": "if0", "linkedVpnTunnel": "..."}]
		field.JSON("interfaces_json", json.RawMessage{}).
			Optional(),

		// NatsJSON contains Cloud NAT configurations.
		//
		//	[{"name": "nat1", "natIpAllocateOption": "AUTO_ONLY"}]
		field.JSON("nats_json", json.RawMessage{}).
			Optional(),

		field.Bool("encrypted_interconnect_router").
			Default(false),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeRouter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeRouter) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_routers"},
	}
}
