package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeComputeServerGroup stores historical snapshots of GreenNode server groups.
type BronzeHistoryGreenNodeComputeServerGroup struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeComputeServerGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeComputeServerGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("policy_id").
			Optional(),
		field.String("policy_name").
			Optional(),
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeComputeServerGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeComputeServerGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_server_groups_history"},
	}
}

// BronzeHistoryGreenNodeComputeServerGroupMember stores historical snapshots of server group members.
type BronzeHistoryGreenNodeComputeServerGroupMember struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeComputeServerGroupMember) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("server_group_history_id").
			Comment("Links to parent BronzeHistoryGreenNodeComputeServerGroup"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),
		field.String("uuid").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeComputeServerGroupMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("server_group_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGreenNodeComputeServerGroupMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_server_group_members_history"},
	}
}
