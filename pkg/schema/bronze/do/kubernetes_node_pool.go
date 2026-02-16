package do

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDOKubernetesNodePool represents a DigitalOcean Kubernetes node pool in the bronze layer.
type BronzeDOKubernetesNodePool struct {
	ent.Schema
}

func (BronzeDOKubernetesNodePool) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDOKubernetesNodePool) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite ID: {clusterID}:{nodePoolID}"),
		field.String("cluster_id").
			NotEmpty(),
		field.String("node_pool_id").
			NotEmpty().
			Comment("Actual node pool UUID"),
		field.String("name").
			Optional(),
		field.String("size").
			Optional().
			Comment("Droplet size slug"),
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
			Optional().
			Comment("Workload placement (security signal)"),
		field.JSON("taints_json", json.RawMessage{}).
			Optional().
			Comment("Workload isolation (security signal)"),
		field.JSON("nodes_json", json.RawMessage{}).
			Optional().
			Comment("Embedded node instances"),
	}
}

func (BronzeDOKubernetesNodePool) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_id"),
		index.Fields("size"),
		index.Fields("collected_at"),
	}
}

func (BronzeDOKubernetesNodePool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_kubernetes_node_pools"},
	}
}
