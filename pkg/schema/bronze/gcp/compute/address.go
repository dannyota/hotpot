package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeAddress represents a GCP Compute Engine regional address in the bronze layer.
// Fields preserve raw API response data from compute.addresses.aggregatedList.
type BronzeGCPComputeAddress struct {
	ent.Schema
}

func (BronzeGCPComputeAddress) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeAddress) Fields() []ent.Field {
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
		field.String("address").
			Optional(),
		field.String("address_type").
			Optional(),
		field.String("ip_version").
			Optional(),
		field.String("ipv6_endpoint_type").
			Optional(),
		field.String("ip_collection").
			Optional(),
		field.String("region").
			Optional(),
		field.String("status").
			Optional(),
		field.String("purpose").
			Optional(),
		field.String("network").
			Optional(),
		field.String("subnetwork").
			Optional(),
		field.String("network_tier").
			Optional(),
		field.Int32("prefix_length").
			Default(0),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("label_fingerprint").
			Optional(),

		// UsersJSON contains list of resource URLs using this address.
		//
		//	["projects/.../instances/vm1", "projects/.../forwardingRules/rule1"]
		field.JSON("users_json", json.RawMessage{}).
			Optional().
			Comment("List of resource URLs using this address"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeAddress) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPComputeAddressLabel.Type),
	}
}

func (BronzeGCPComputeAddress) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("status"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeAddress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_addresses"},
	}
}

// BronzeGCPComputeAddressLabel represents a label attached to a GCP Compute address.
type BronzeGCPComputeAddressLabel struct {
	ent.Schema
}

func (BronzeGCPComputeAddressLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPComputeAddressLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("address", BronzeGCPComputeAddress.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeAddressLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_address_labels"},
	}
}

// BronzeGCPComputeGlobalAddress represents a GCP Compute Engine global address in the bronze layer.
// Fields preserve raw API response data from compute.globalAddresses.list.
type BronzeGCPComputeGlobalAddress struct {
	ent.Schema
}

func (BronzeGCPComputeGlobalAddress) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeGlobalAddress) Fields() []ent.Field {
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
		field.String("address").
			Optional(),
		field.String("address_type").
			Optional(),
		field.String("ip_version").
			Optional(),
		field.String("ipv6_endpoint_type").
			Optional(),
		field.String("ip_collection").
			Optional(),
		field.String("region").
			Optional(),
		field.String("status").
			Optional(),
		field.String("purpose").
			Optional(),
		field.String("network").
			Optional(),
		field.String("subnetwork").
			Optional(),
		field.String("network_tier").
			Optional(),
		field.Int32("prefix_length").
			Default(0),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("label_fingerprint").
			Optional(),

		// UsersJSON contains list of resource URLs using this address.
		//
		//	["projects/.../forwardingRules/rule1", "projects/.../targetPools/pool1"]
		field.JSON("users_json", json.RawMessage{}).
			Optional().
			Comment("List of resource URLs using this address"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeGlobalAddress) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPComputeGlobalAddressLabel.Type),
	}
}

func (BronzeGCPComputeGlobalAddress) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("status"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeGlobalAddress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_global_addresses"},
	}
}

// BronzeGCPComputeGlobalAddressLabel represents a label attached to a GCP Compute global address.
type BronzeGCPComputeGlobalAddressLabel struct {
	ent.Schema
}

func (BronzeGCPComputeGlobalAddressLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPComputeGlobalAddressLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("global_address", BronzeGCPComputeGlobalAddress.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeGlobalAddressLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_global_address_labels"},
	}
}
