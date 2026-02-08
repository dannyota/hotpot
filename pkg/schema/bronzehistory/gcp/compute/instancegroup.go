package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeInstanceGroup stores historical snapshots of GCP Compute instance groups.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeInstanceGroup struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeInstanceGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze instance group by resource_id"),

		// All instance group fields (same as bronze.BronzeGCPComputeInstanceGroup)
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
			Optional(),
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

func (BronzeHistoryGCPComputeInstanceGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeInstanceGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_instance_groups_history"},
	}
}

// BronzeHistoryGCPComputeInstanceGroupNamedPort stores historical snapshots of instance group named ports.
// Links via group_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeInstanceGroupNamedPort struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceGroupNamedPort) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("group_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeInstanceGroup"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Named port fields
		field.String("name").
			NotEmpty(),
		field.Int32("port"),
	}
}

func (BronzeHistoryGCPComputeInstanceGroupNamedPort) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeInstanceGroupNamedPort) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_instance_group_named_ports_history"},
	}
}

// BronzeHistoryGCPComputeInstanceGroupMember stores historical snapshots of instance group members.
// Links via group_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeInstanceGroupMember struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceGroupMember) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("group_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeInstanceGroup"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Member fields
		field.String("instance_url").
			NotEmpty(),
		field.String("instance_name").
			Optional(),
		field.String("status").
			Optional(),
	}
}

func (BronzeHistoryGCPComputeInstanceGroupMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeInstanceGroupMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_instance_group_members_history"},
	}
}
