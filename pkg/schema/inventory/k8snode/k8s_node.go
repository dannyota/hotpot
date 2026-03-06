package k8snode

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	inventorymixin "danny.vn/hotpot/pkg/schema/inventory/mixin"
)

// InventoryK8sNode is the final merged Kubernetes node table.
type InventoryK8sNode struct {
	ent.Schema
}

func (InventoryK8sNode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		inventorymixin.Timestamp{},
	}
}

func (InventoryK8sNode) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("node_name").NotEmpty(),
		field.String("cluster_name").NotEmpty(),
		field.String("node_pool"),
		field.String("status"),
		field.String("provisioning").Optional(),
		field.String("cloud_project").Optional(),
		field.String("cloud_zone").Optional(),
		field.String("cloud_machine_type").Optional(),
		field.String("internal_ip").Optional(),
		field.String("external_ip").Optional(),
	}
}

func (InventoryK8sNode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bronze_links", InventoryK8sNodeBronzeLink.Type),
	}
}

func (InventoryK8sNode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_name"),
		index.Fields("collected_at"),
	}
}

func (InventoryK8sNode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "k8s_nodes"},
	}
}
