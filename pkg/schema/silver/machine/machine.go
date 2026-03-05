package machine

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	silvermixin "github.com/dannyota/hotpot/pkg/schema/silver/mixin"
)

// SilverMachine is the final merged machine table.
type SilverMachine struct {
	ent.Schema
}

func (SilverMachine) Mixin() []ent.Mixin {
	return []ent.Mixin{
		silvermixin.Timestamp{},
	}
}

func (SilverMachine) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("hostname").NotEmpty(),
		field.String("os_type"),
		field.String("os_name").Optional(),
		field.String("status"),
		field.String("internal_ip").Optional(),
		field.String("external_ip").Optional(),
		field.String("environment").Optional(),
		field.String("cloud_project").Optional(),
		field.String("cloud_zone").Optional(),
		field.String("cloud_machine_type").Optional(),
	}
}

func (SilverMachine) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bronze_links", SilverMachineBronzeLink.Type),
	}
}

func (SilverMachine) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("os_type"),
		index.Fields("environment"),
		index.Fields("collected_at"),
	}
}

func (SilverMachine) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "machines"},
	}
}
