package container

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPContainerCluster stores historical snapshots of GKE clusters.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPContainerCluster struct {
	ent.Schema
}

func (BronzeHistoryGCPContainerCluster) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPContainerCluster) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze cluster by resource_id"),

		// All cluster fields (same as bronze.BronzeGCPContainerCluster)
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

func (BronzeHistoryGCPContainerCluster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPContainerCluster) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_container_clusters_history"},
	}
}

// BronzeHistoryGCPContainerClusterAddon stores historical snapshots of cluster addons.
// Links via cluster_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPContainerClusterAddon struct {
	ent.Schema
}

func (BronzeHistoryGCPContainerClusterAddon) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("cluster_history_id").
			Comment("Links to parent BronzeHistoryGCPContainerCluster"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Addon fields
		field.String("addon_name").
			Optional(),
		field.Bool("enabled").
			Default(false),
		field.JSON("config_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGCPContainerClusterAddon) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPContainerClusterAddon) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_container_cluster_addons_history"},
	}
}

// BronzeHistoryGCPContainerClusterCondition stores historical snapshots of cluster conditions.
// Links via cluster_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPContainerClusterCondition struct {
	ent.Schema
}

func (BronzeHistoryGCPContainerClusterCondition) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("cluster_history_id").
			Comment("Links to parent BronzeHistoryGCPContainerCluster"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Condition fields
		field.String("code").
			Optional(),
		field.String("message").
			Optional(),
		field.String("canonical_code").
			Optional(),
	}
}

func (BronzeHistoryGCPContainerClusterCondition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPContainerClusterCondition) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_container_cluster_conditions_history"},
	}
}

// BronzeHistoryGCPContainerClusterLabel stores historical snapshots of cluster labels.
// Links via cluster_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPContainerClusterLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPContainerClusterLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("cluster_history_id").
			Comment("Links to parent BronzeHistoryGCPContainerCluster"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Label fields
		field.String("key").
			Optional(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPContainerClusterLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPContainerClusterLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_container_cluster_labels_history"},
	}
}

// BronzeHistoryGCPContainerClusterNodePool stores historical snapshots of cluster node pools.
// Links via cluster_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPContainerClusterNodePool struct {
	ent.Schema
}

func (BronzeHistoryGCPContainerClusterNodePool) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("cluster_history_id").
			Comment("Links to parent BronzeHistoryGCPContainerCluster"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// All node pool fields (same as bronze.BronzeGCPContainerClusterNodePool)
		field.String("name").
			Optional(),
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

		// JSONB fields
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

func (BronzeHistoryGCPContainerClusterNodePool) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPContainerClusterNodePool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_container_cluster_node_pools_history"},
	}
}
