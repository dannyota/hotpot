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

// BronzeDOKubernetesCluster represents a DigitalOcean Kubernetes cluster in the bronze layer.
type BronzeDOKubernetesCluster struct {
	ent.Schema
}

func (BronzeDOKubernetesCluster) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDOKubernetesCluster) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("DigitalOcean Kubernetes Cluster UUID"),
		field.String("name").
			Optional(),
		field.String("region_slug").
			Optional(),
		field.String("version_slug").
			Optional().
			Comment("Kubernetes version (security signal)"),
		field.String("cluster_subnet").
			Optional(),
		field.String("service_subnet").
			Optional(),
		field.String("ipv4").
			Optional(),
		field.String("endpoint").
			Optional().
			Comment("API server URL"),
		field.String("vpc_uuid").
			Optional(),
		field.Bool("ha").
			Default(false).
			Comment("HA control plane (security signal)"),
		field.Bool("auto_upgrade").
			Default(false).
			Comment("Automatic patching (security signal)"),
		field.Bool("surge_upgrade").
			Default(false).
			Comment("Zero-downtime upgrades (security signal)"),
		field.Bool("registry_enabled").
			Default(false),
		field.String("status_state").
			Optional(),
		field.String("status_message").
			Optional(),
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),
		field.JSON("maintenance_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("control_plane_firewall_json", json.RawMessage{}).
			Optional().
			Comment("API access control (security signal)"),
		field.JSON("autoscaler_config_json", json.RawMessage{}).
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
		field.Time("api_updated_at").
			Optional().
			Nillable(),
	}
}

func (BronzeDOKubernetesCluster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("version_slug"),
		index.Fields("region_slug"),
		index.Fields("status_state"),
		index.Fields("vpc_uuid"),
		index.Fields("ha"),
		index.Fields("collected_at"),
	}
}

func (BronzeDOKubernetesCluster) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_kubernetes_clusters"},
	}
}
