package do

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDOKubernetesNodePool stores historical snapshots of DigitalOcean Kubernetes node pools.
type BronzeHistoryDOKubernetesNodePool struct {
	ent.Schema
}

func (BronzeHistoryDOKubernetesNodePool) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDOKubernetesNodePool) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze KubernetesNodePool by resource_id"),
		field.String("cluster_id").
			NotEmpty(),
		field.String("node_pool_id").
			NotEmpty(),
		field.String("name").
			Optional(),
		field.String("size").
			Optional(),
		field.Int("count").
			Default(0),
		field.Bool("auto_scale").
			Default(false),
		field.Int("min_nodes").
			Default(0),
		field.Int("max_nodes").
			Default(0),
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),
		field.JSON("taints_json", json.RawMessage{}).
			Optional(),
		field.JSON("nodes_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryDOKubernetesNodePool) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("cluster_id"),
		index.Fields("size"),
	}
}

func (BronzeHistoryDOKubernetesNodePool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_kubernetes_node_pools_history"},
	}
}
