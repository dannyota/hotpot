package k8snode

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// InventoryK8sNodeBronzeLink tracks which bronze records contributed to a k8s node.
type InventoryK8sNodeBronzeLink struct {
	ent.Schema
}

func (InventoryK8sNodeBronzeLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").NotEmpty(),
		field.String("bronze_table").NotEmpty(),
		field.String("bronze_resource_id").NotEmpty(),
	}
}

func (InventoryK8sNodeBronzeLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("k8s_node", InventoryK8sNode.Type).Ref("bronze_links").Unique().Required(),
	}
}

func (InventoryK8sNodeBronzeLink) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_k8s_node_links"},
	}
}
