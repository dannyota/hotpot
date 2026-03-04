package k8snode

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SilverK8sNodeBronzeLink tracks which bronze records contributed to a k8s node.
type SilverK8sNodeBronzeLink struct {
	ent.Schema
}

func (SilverK8sNodeBronzeLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").NotEmpty(),
		field.String("bronze_table").NotEmpty(),
		field.String("bronze_resource_id").NotEmpty(),
	}
}

func (SilverK8sNodeBronzeLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("k8s_node", SilverK8sNode.Type).Ref("bronze_links").Unique().Required(),
	}
}

func (SilverK8sNodeBronzeLink) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "k8s_node_bronze_links"},
	}
}
