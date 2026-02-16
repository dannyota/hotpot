package redis

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPRedisInstance represents a GCP Memorystore Redis instance in the bronze layer.
// Fields preserve raw API response data from redis.projects.locations.instances.list.
type BronzeGCPRedisInstance struct {
	ent.Schema
}

func (BronzeGCPRedisInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPRedisInstance) Fields() []ent.Field {
	return []ent.Field{
		// Primary key - instance resource name (projects/{project}/locations/{location}/instances/{instance})
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Instance resource name"),
		field.String("name").
			NotEmpty(),
		field.String("display_name").
			Optional(),

		// LabelsJSON contains resource labels as key-value pairs.
		//
		//	{"env": "prod", "team": "infra"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		field.String("location_id").
			Optional().
			Comment("Zone where the primary node is provisioned"),
		field.String("alternative_location_id").
			Optional().
			Comment("Zone for additional node (standard tier)"),
		field.String("redis_version").
			Optional().
			Comment("Redis software version (e.g. REDIS_6_X)"),
		field.String("reserved_ip_range").
			Optional().
			Comment("CIDR range of internal addresses reserved for this instance"),
		field.String("secondary_ip_range").
			Optional().
			Comment("Additional IP range for node placement"),
		field.String("host").
			Optional().
			Comment("Hostname or IP of the exposed Redis endpoint"),
		field.Int32("port").
			Optional().
			Comment("Port number of the exposed Redis endpoint"),
		field.String("current_location_id").
			Optional().
			Comment("Current zone where the Redis primary node is located"),
		field.String("create_time").
			Optional(),
		field.Int32("state").
			Optional().
			Comment("Current state (Instance_State enum)"),
		field.String("status_message").
			Optional(),

		// RedisConfigsJSON contains Redis configuration parameters.
		//
		//	{"maxmemory-policy": "volatile-lru", "notify-keyspace-events": ""}
		field.JSON("redis_configs_json", json.RawMessage{}).
			Optional(),

		field.Int32("tier").
			Optional().
			Comment("Service tier (0=BASIC, 1=STANDARD_HA)"),
		field.Int32("memory_size_gb").
			Optional().
			Comment("Redis memory size in GiB"),
		field.String("authorized_network").
			Optional().
			Comment("Full name of the VPC network the instance is connected to"),
		field.String("persistence_iam_identity").
			Optional().
			Comment("IAM identity used by import/export operations"),
		field.Int32("connect_mode").
			Optional().
			Comment("Network connect mode (Instance_ConnectMode enum)"),
		field.Bool("auth_enabled").
			Default(false).
			Comment("Whether OSS Redis AUTH is enabled"),

		// ServerCaCertsJSON contains server CA certificates.
		//
		//	[{"serialNumber": "...", "cert": "...", "createTime": "...", "expireTime": "...", "sha1Fingerprint": "..."}]
		field.JSON("server_ca_certs_json", json.RawMessage{}).
			Optional(),

		field.Int32("transit_encryption_mode").
			Optional().
			Comment("TLS mode (Instance_TransitEncryptionMode enum)"),

		// MaintenancePolicyJSON contains the maintenance policy configuration.
		//
		//	{"createTime": "...", "updateTime": "...", "weeklyMaintenanceWindow": [...]}
		field.JSON("maintenance_policy_json", json.RawMessage{}).
			Optional(),

		// MaintenanceScheduleJSON contains upcoming maintenance event details.
		//
		//	{"startTime": "...", "endTime": "...", "canReschedule": true, "scheduleDeadlineTime": "..."}
		field.JSON("maintenance_schedule_json", json.RawMessage{}).
			Optional(),

		field.Int32("replica_count").
			Optional().
			Comment("Number of replica nodes"),

		// NodesJSON contains info per node.
		//
		//	[{"id": "node-0", "zone": "us-central1-a"}]
		field.JSON("nodes_json", json.RawMessage{}).
			Optional(),

		field.String("read_endpoint").
			Optional().
			Comment("Hostname or IP of the readonly Redis endpoint"),
		field.Int32("read_endpoint_port").
			Optional().
			Comment("Port number of the readonly Redis endpoint"),
		field.Int32("read_replicas_mode").
			Optional().
			Comment("Read replicas mode (Instance_ReadReplicasMode enum)"),
		field.String("customer_managed_key").
			Optional().
			Comment("KMS key reference for CMEK"),

		// PersistenceConfigJSON contains persistence configuration.
		//
		//	{"persistenceMode": "RDB", "rdbSnapshotPeriod": "TWELVE_HOURS", "rdbSnapshotStartTime": "...", "rdbNextSnapshotTime": "..."}
		field.JSON("persistence_config_json", json.RawMessage{}).
			Optional(),

		// SuspensionReasonsJSON contains reasons the instance is in SUSPENDED state.
		//
		//	[0, 1]
		field.JSON("suspension_reasons_json", json.RawMessage{}).
			Optional(),

		field.String("maintenance_version").
			Optional().
			Comment("Self service update maintenance version"),

		// AvailableMaintenanceVersionsJSON contains available maintenance versions.
		//
		//	["20210712_00_00", "20210801_00_00"]
		field.JSON("available_maintenance_versions_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPRedisInstance) Edges() []ent.Edge {
	return nil
}

func (BronzeGCPRedisInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPRedisInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_redis_instances"},
	}
}
