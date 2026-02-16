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

// BronzeHistoryDOKubernetesCluster stores historical snapshots of DigitalOcean Kubernetes clusters.
type BronzeHistoryDOKubernetesCluster struct {
	ent.Schema
}

func (BronzeHistoryDOKubernetesCluster) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDOKubernetesCluster) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze KubernetesCluster by resource_id"),
		field.String("name").
			Optional(),
		field.String("region_slug").
			Optional(),
		field.String("version_slug").
			Optional(),
		field.String("cluster_subnet").
			Optional(),
		field.String("service_subnet").
			Optional(),
		field.String("ipv4").
			Optional(),
		field.String("endpoint").
			Optional(),
		field.String("vpc_uuid").
			Optional(),
		field.Bool("ha").
			Default(false),
		field.Bool("auto_upgrade").
			Default(false),
		field.Bool("surge_upgrade").
			Default(false),
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
			Optional(),
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

func (BronzeHistoryDOKubernetesCluster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("version_slug"),
		index.Fields("region_slug"),
		index.Fields("status_state"),
		index.Fields("vpc_uuid"),
		index.Fields("ha"),
	}
}

func (BronzeHistoryDOKubernetesCluster) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_kubernetes_clusters_history"},
	}
}
