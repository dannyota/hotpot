package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeInstanceGroup represents a GCP Compute Engine instance group in the bronze layer.
// Fields preserve raw API response data from compute.instanceGroups.aggregatedList.
type BronzeGCPComputeInstanceGroup struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeInstanceGroup) Fields() []ent.Field {
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
		field.String("zone").
			Optional(),
		field.String("network").
			Optional(),
		field.String("subnetwork").
			Optional(),
		field.Int32("size").
			Optional().
			Comment("Number of instances in the group"),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("fingerprint").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeInstanceGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("named_ports", BronzeGCPComputeInstanceGroupNamedPort.Type),
		edge.To("members", BronzeGCPComputeInstanceGroupMember.Type),
	}
}

func (BronzeGCPComputeInstanceGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeInstanceGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_groups"},
	}
}

// BronzeGCPComputeInstanceGroupNamedPort represents a named port on a GCP instance group.
type BronzeGCPComputeInstanceGroupNamedPort struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceGroupNamedPort) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),
		field.Int32("port"),
	}
}

func (BronzeGCPComputeInstanceGroupNamedPort) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance_group", BronzeGCPComputeInstanceGroup.Type).
			Ref("named_ports").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeInstanceGroupNamedPort) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_group_named_ports"},
	}
}

// BronzeGCPComputeInstanceGroupMember represents a member instance in a GCP instance group.
type BronzeGCPComputeInstanceGroupMember struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceGroupMember) Fields() []ent.Field {
	return []ent.Field{
		field.String("instance_url").
			NotEmpty().
			Comment("URL of the instance"),
		field.String("instance_name").
			Optional(),
		field.String("status").
			Optional(),
	}
}

func (BronzeGCPComputeInstanceGroupMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance_group", BronzeGCPComputeInstanceGroup.Type).
			Ref("members").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeInstanceGroupMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_group_members"},
	}
}
