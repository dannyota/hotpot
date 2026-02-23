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

// BronzeGreenNodeComputeServerGroup represents a GreenNode server group in the bronze layer.
type BronzeGreenNodeComputeServerGroup struct {
	ent.Schema
}

func (BronzeGreenNodeComputeServerGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeComputeServerGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Server group UUID"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("policy_id").
			Optional(),
		field.String("policy_name").
			Optional(),
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeComputeServerGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", BronzeGreenNodeComputeServerGroupMember.Type),
	}
}

func (BronzeGreenNodeComputeServerGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeComputeServerGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_server_groups"},
	}
}

// BronzeGreenNodeComputeServerGroupMember represents a server in a server group.
type BronzeGreenNodeComputeServerGroupMember struct {
	ent.Schema
}

func (BronzeGreenNodeComputeServerGroupMember) Fields() []ent.Field {
	return []ent.Field{
		field.String("uuid").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
	}
}

func (BronzeGreenNodeComputeServerGroupMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("server_group", BronzeGreenNodeComputeServerGroup.Type).
			Ref("members").
			Unique().
			Required(),
	}
}

func (BronzeGreenNodeComputeServerGroupMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_server_group_members"},
	}
}
