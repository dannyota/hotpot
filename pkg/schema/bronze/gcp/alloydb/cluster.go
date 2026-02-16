package alloydb

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPAlloyDBCluster represents a GCP AlloyDB cluster in the bronze layer.
type BronzeGCPAlloyDBCluster struct {
	ent.Schema
}

func (BronzeGCPAlloyDBCluster) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPAlloyDBCluster) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Cluster resource name (projects/{project}/locations/{location}/clusters/{cluster})"),
		field.String("name").
			NotEmpty(),
		field.String("display_name").
			Optional(),
		field.String("uid").
			Optional().
			Comment("System-generated UID"),
		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),
		field.String("delete_time").
			Optional(),

		// LabelsJSON contains user-defined labels.
		//
		//	{"env": "prod", "team": "data"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		field.Int("state").
			Optional().
			Comment("Cluster state (0=UNSPECIFIED, 1=READY, 2=STOPPED, 3=EMPTY, 4=CREATING, 5=DELETING, 6=FAILED, 7=BOOTSTRAPPING, 8=MAINTENANCE, 9=PROMOTING)"),

		field.Int("cluster_type").
			Optional().
			Comment("Cluster type (0=UNSPECIFIED, 1=PRIMARY, 2=SECONDARY)"),

		field.Int("database_version").
			Optional().
			Comment("Database engine major version (0=UNSPECIFIED, 1=POSTGRES_13, 2=POSTGRES_14, 3=POSTGRES_15, 4=POSTGRES_16)"),

		// NetworkConfigJSON contains the network configuration for the cluster.
		//
		//	{"network": "projects/.../global/networks/...", "allocatedIpRange": "..."}
		field.JSON("network_config_json", json.RawMessage{}).
			Optional(),

		field.String("network").
			Optional().
			Comment("Deprecated: VPC network resource link, use network_config_json instead"),

		field.String("etag").
			Optional(),

		// AnnotationsJSON contains user-managed annotations.
		//
		//	{"key1": "value1", "key2": "value2"}
		field.JSON("annotations_json", json.RawMessage{}).
			Optional(),

		field.Bool("reconciling").
			Optional().
			Comment("Whether the cluster is being reconciled"),

		// InitialUserJSON contains the initial user configuration.
		//
		//	{"user": "admin", "password": "..."}
		field.JSON("initial_user_json", json.RawMessage{}).
			Optional(),

		// AutomatedBackupPolicyJSON contains the automated backup policy.
		//
		//	{"weeklySchedule": {...}, "backupWindow": "...", "enabled": true, ...}
		field.JSON("automated_backup_policy_json", json.RawMessage{}).
			Optional(),

		// SslConfigJSON contains SSL configuration (deprecated, use encryption).
		//
		//	{"sslMode": "...", "caSource": "..."}
		field.JSON("ssl_config_json", json.RawMessage{}).
			Optional(),

		// EncryptionConfigJSON contains CMEK encryption configuration.
		//
		//	{"kmsKeyName": "projects/.../locations/.../keyRings/.../cryptoKeys/..."}
		field.JSON("encryption_config_json", json.RawMessage{}).
			Optional(),

		// EncryptionInfoJSON contains encryption status information.
		//
		//	{"encryptionType": "...", "kmsKeyVersions": [...]}
		field.JSON("encryption_info_json", json.RawMessage{}).
			Optional(),

		// ContinuousBackupConfigJSON contains continuous backup configuration.
		//
		//	{"enabled": true, "recoveryWindowDays": 14, "encryptionConfig": {...}}
		field.JSON("continuous_backup_config_json", json.RawMessage{}).
			Optional(),

		// ContinuousBackupInfoJSON contains continuous backup status.
		//
		//	{"encryptionInfo": {...}, "enabledTime": "...", "schedule": [...]}
		field.JSON("continuous_backup_info_json", json.RawMessage{}).
			Optional(),

		// SecondaryConfigJSON contains cross-region replication config for SECONDARY clusters.
		//
		//	{"primaryClusterName": "projects/.../locations/.../clusters/..."}
		field.JSON("secondary_config_json", json.RawMessage{}).
			Optional(),

		// PrimaryConfigJSON contains cross-region replication config for PRIMARY clusters.
		//
		//	{"secondaryClusterNames": ["projects/.../locations/.../clusters/..."]}
		field.JSON("primary_config_json", json.RawMessage{}).
			Optional(),

		field.Bool("satisfies_pzs").
			Optional().
			Comment("Reserved for future use"),

		// PscConfigJSON contains Private Service Connect configuration.
		//
		//	{"pscEnabled": true}
		field.JSON("psc_config_json", json.RawMessage{}).
			Optional(),

		// MaintenanceUpdatePolicyJSON contains the maintenance update policy.
		//
		//	{"maintenanceWindows": [{"day": "MONDAY", "startTime": {...}}]}
		field.JSON("maintenance_update_policy_json", json.RawMessage{}).
			Optional(),

		// MaintenanceScheduleJSON contains the generated maintenance schedule.
		//
		//	{"startTime": "...", "endTime": "..."}
		field.JSON("maintenance_schedule_json", json.RawMessage{}).
			Optional(),

		field.Int("subscription_type").
			Optional().
			Comment("Subscription type (0=UNSPECIFIED, 1=STANDARD, 2=TRIAL)"),

		// TrialMetadataJSON contains metadata for free trial clusters.
		//
		//	{"startTime": "...", "endTime": "...", "upgradeTime": "..."}
		field.JSON("trial_metadata_json", json.RawMessage{}).
			Optional(),

		// TagsJSON contains tag keys/values directly bound to this resource.
		//
		//	{"123/environment": "production", "123/costCenter": "marketing"}
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			NotEmpty(),
	}
}

func (BronzeGCPAlloyDBCluster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPAlloyDBCluster) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_alloydb_clusters"},
	}
}
