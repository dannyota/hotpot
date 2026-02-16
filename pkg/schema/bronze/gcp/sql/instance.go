package sql

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

// BronzeGCPSQLInstance represents a GCP Cloud SQL instance in the bronze layer.
// Fields preserve raw API response data from sqladmin.instances.list.
type BronzeGCPSQLInstance struct {
	ent.Schema
}

func (BronzeGCPSQLInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPSQLInstance) Fields() []ent.Field {
	return []ent.Field{
		// Primary key - use instance name as resource_id (SQL instances use name as identifier)
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Instance name, used as primary key for linking"),
		field.String("name").
			NotEmpty(),
		field.String("database_version").
			Optional().
			Comment("Database engine version (e.g. MYSQL_8_0, POSTGRES_15)"),
		field.String("state").
			Optional().
			Comment("Instance state: RUNNABLE, SUSPENDED, PENDING_DELETE, etc."),
		field.String("region").
			Optional(),
		field.String("gce_zone").
			Optional(),
		field.String("secondary_gce_zone").
			Optional(),
		field.String("instance_type").
			Optional().
			Comment("CLOUD_SQL_INSTANCE, ON_PREMISES_INSTANCE, READ_REPLICA_INSTANCE"),
		field.String("connection_name").
			Optional().
			Comment("Connection name for connecting: project:region:instance"),
		field.String("service_account_email_address").
			Optional(),
		field.String("self_link").
			Optional(),

		// SettingsJSON contains instance settings configuration.
		//
		//	{
		//	  "tier": "db-custom-2-7680",
		//	  "activationPolicy": "ALWAYS",
		//	  "availabilityType": "REGIONAL",
		//	  "dataDiskSizeGb": "100",
		//	  "dataDiskType": "PD_SSD",
		//	  "backupConfiguration": {...},
		//	  "ipConfiguration": {...},
		//	  "databaseFlags": [...]
		//	}
		field.JSON("settings_json", json.RawMessage{}).
			Optional(),

		// ServerCaCertJSON contains the SSL CA certificate for the instance.
		//
		//	{
		//	  "kind": "sql#sslCert",
		//	  "certSerialNumber": "...",
		//	  "cert": "-----BEGIN CERTIFICATE-----...",
		//	  "commonName": "...",
		//	  "sha1Fingerprint": "...",
		//	  "expirationTime": "2030-..."
		//	}
		field.JSON("server_ca_cert_json", json.RawMessage{}).
			Optional(),

		// IpAddressesJSON contains IP addresses assigned to the instance.
		//
		//	[
		//	  {"type": "PRIMARY", "ipAddress": "10.0.0.1"},
		//	  {"type": "PRIVATE", "ipAddress": "10.0.0.2"}
		//	]
		field.JSON("ip_addresses_json", json.RawMessage{}).
			Optional(),

		// ReplicaConfigurationJSON contains read replica configuration.
		//
		//	{
		//	  "mysqlReplicaConfiguration": {...},
		//	  "failoverTarget": false
		//	}
		field.JSON("replica_configuration_json", json.RawMessage{}).
			Optional(),

		// FailoverReplicaJSON contains failover replica information.
		//
		//	{
		//	  "name": "replica-name",
		//	  "available": true
		//	}
		field.JSON("failover_replica_json", json.RawMessage{}).
			Optional(),

		// DiskEncryptionConfigurationJSON contains CMEK configuration.
		//
		//	{
		//	  "kmsKeyName": "projects/.../cryptoKeys/...",
		//	  "kind": "sql#diskEncryptionConfiguration"
		//	}
		field.JSON("disk_encryption_configuration_json", json.RawMessage{}).
			Optional(),

		// DiskEncryptionStatusJSON contains CMEK status.
		//
		//	{
		//	  "kmsKeyVersionName": "projects/.../cryptoKeyVersions/1",
		//	  "kind": "sql#diskEncryptionStatus"
		//	}
		field.JSON("disk_encryption_status_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPSQLInstance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPSQLInstanceLabel.Type),
	}
}

func (BronzeGCPSQLInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("state"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPSQLInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_sql_instances"},
	}
}

// BronzeGCPSQLInstanceLabel represents a label on a GCP Cloud SQL instance.
// Data from instance.settings.userLabels map.
type BronzeGCPSQLInstanceLabel struct {
	ent.Schema
}

func (BronzeGCPSQLInstanceLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value").
			Optional(),
	}
}

func (BronzeGCPSQLInstanceLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", BronzeGCPSQLInstance.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPSQLInstanceLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_sql_instance_labels"},
	}
}
