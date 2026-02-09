package container

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPContainerCluster represents a GKE cluster in the bronze layer.
// Fields preserve raw API response data from container.clusters.list.
type BronzeGCPContainerCluster struct {
	ent.Schema
}

func (BronzeGCPContainerCluster) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPContainerCluster) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields (preserving original API field structure)
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID, used as primary key for linking"),
		field.String("name").
			NotEmpty(),
		field.String("location").
			Optional(),
		field.String("zone").
			Optional(),
		field.String("description").
			Optional(),
		field.String("initial_cluster_version").
			Optional(),
		field.String("current_master_version").
			Optional(),
		field.String("current_node_version").
			Optional(),
		field.String("status").
			Optional(),
		field.String("status_message").
			Optional(),
		field.Int32("current_node_count").
			Optional(),
		field.String("network").
			Optional(),
		field.String("subnetwork").
			Optional(),
		field.String("cluster_ipv4_cidr").
			Optional(),
		field.String("services_ipv4_cidr").
			Optional(),
		field.Int32("node_ipv4_cidr_size").
			Optional(),
		field.String("endpoint").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("create_time").
			Optional(),
		field.String("expire_time").
			Optional(),
		field.String("etag").
			Optional(),
		field.String("label_fingerprint").
			Optional(),
		field.String("logging_service").
			Optional(),
		field.String("monitoring_service").
			Optional(),
		field.Bool("enable_kubernetes_alpha").
			Default(false),
		field.Bool("enable_tpu").
			Default(false),
		field.String("tpu_ipv4_cidr_block").
			Optional(),

		// Nested objects stored as JSONB
		field.JSON("addons_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("private_cluster_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("ip_allocation_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("network_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("master_auth_json", json.RawMessage{}).
			Optional(),
		field.JSON("autoscaling_json", json.RawMessage{}).
			Optional(),
		field.JSON("vertical_pod_autoscaling_json", json.RawMessage{}).
			Optional(),
		field.JSON("monitoring_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("logging_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("maintenance_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("database_encryption_json", json.RawMessage{}).
			Optional(),
		field.JSON("workload_identity_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("autopilot_json", json.RawMessage{}).
			Optional(),
		field.JSON("release_channel_json", json.RawMessage{}).
			Optional(),
		field.JSON("binary_authorization_json", json.RawMessage{}).
			Optional(),
		field.JSON("security_posture_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("node_pool_defaults_json", json.RawMessage{}).
			Optional(),
		field.JSON("fleet_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPContainerCluster) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPContainerClusterLabel.Type),
		edge.To("addons", BronzeGCPContainerClusterAddon.Type),
		edge.To("conditions", BronzeGCPContainerClusterCondition.Type),
		edge.To("node_pools", BronzeGCPContainerClusterNodePool.Type),
	}
}

func (BronzeGCPContainerCluster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPContainerCluster) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_container_clusters"},
	}
}

// BronzeGCPContainerClusterAddon represents an addon configuration for a cluster.
// Data from cluster.addonsConfig, one row per addon.
type BronzeGCPContainerClusterAddon struct {
	ent.Schema
}

func (BronzeGCPContainerClusterAddon) Fields() []ent.Field {
	return []ent.Field{
		field.String("addon_name").
			NotEmpty(),
		field.Bool("enabled").
			Default(false),
		field.JSON("config_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeGCPContainerClusterAddon) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cluster", BronzeGCPContainerCluster.Type).
			Ref("addons").
			Unique().
			Required(),
	}
}

func (BronzeGCPContainerClusterAddon) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_container_cluster_addons"},
	}
}

// BronzeGCPContainerClusterCondition represents a status condition on a cluster.
// Data from cluster.conditions[].
type BronzeGCPContainerClusterCondition struct {
	ent.Schema
}

func (BronzeGCPContainerClusterCondition) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			Optional(),
		field.String("message").
			Optional(),
		field.String("canonical_code").
			Optional(),
	}
}

func (BronzeGCPContainerClusterCondition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cluster", BronzeGCPContainerCluster.Type).
			Ref("conditions").
			Unique().
			Required(),
	}
}

func (BronzeGCPContainerClusterCondition) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_container_cluster_conditions"},
	}
}

// BronzeGCPContainerClusterLabel represents a resource label on a cluster.
// Data from cluster.resourceLabels map.
type BronzeGCPContainerClusterLabel struct {
	ent.Schema
}

func (BronzeGCPContainerClusterLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPContainerClusterLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cluster", BronzeGCPContainerCluster.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPContainerClusterLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_container_cluster_labels"},
	}
}

// BronzeGCPContainerClusterNodePool represents a node pool in a cluster.
// Data from cluster.nodePools[].
type BronzeGCPContainerClusterNodePool struct {
	ent.Schema
}

func (BronzeGCPContainerClusterNodePool) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),
		field.String("version").
			Optional(),
		field.String("status").
			Optional(),
		field.String("status_message").
			Optional(),
		field.Int32("initial_node_count").
			Optional(),
		field.String("self_link").
			Optional(),
		field.Int32("pod_ipv4_cidr_size").
			Optional(),
		field.String("etag").
			Optional(),

		// Nested objects stored as JSONB (includes locations, config with taints/labels)
		field.JSON("locations_json", json.RawMessage{}).
			Optional(),
		field.JSON("config_json", json.RawMessage{}).
			Optional(),
		field.JSON("autoscaling_json", json.RawMessage{}).
			Optional(),
		field.JSON("management_json", json.RawMessage{}).
			Optional(),
		field.JSON("upgrade_settings_json", json.RawMessage{}).
			Optional(),
		field.JSON("network_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("placement_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("max_pods_constraint_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeGCPContainerClusterNodePool) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cluster", BronzeGCPContainerCluster.Type).
			Ref("node_pools").
			Unique().
			Required(),
	}
}

func (BronzeGCPContainerClusterNodePool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_container_cluster_node_pools"},
	}
}
