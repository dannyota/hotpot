package sql

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPSQLInstance stores historical snapshots of GCP Cloud SQL instances.
// Uses resource_id for lookup (instance name), with valid_from/valid_to for time range.
type BronzeHistoryGCPSQLInstance struct {
	ent.Schema
}

func (BronzeHistoryGCPSQLInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPSQLInstance) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze SQL instance by resource_id"),

		// All instance fields (same as bronze.BronzeGCPSQLInstance)
		field.String("name").
			NotEmpty(),
		field.String("database_version").
			Optional(),
		field.String("state").
			Optional(),
		field.String("region").
			Optional(),
		field.String("gce_zone").
			Optional(),
		field.String("secondary_gce_zone").
			Optional(),
		field.String("instance_type").
			Optional(),
		field.String("connection_name").
			Optional(),
		field.String("service_account_email_address").
			Optional(),
		field.String("self_link").
			Optional(),

		// JSONB fields
		field.JSON("settings_json", json.RawMessage{}).
			Optional(),
		field.JSON("server_ca_cert_json", json.RawMessage{}).
			Optional(),
		field.JSON("ip_addresses_json", json.RawMessage{}).
			Optional(),
		field.JSON("replica_configuration_json", json.RawMessage{}).
			Optional(),
		field.JSON("failover_replica_json", json.RawMessage{}).
			Optional(),
		field.JSON("disk_encryption_configuration_json", json.RawMessage{}).
			Optional(),
		field.JSON("disk_encryption_status_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPSQLInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPSQLInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_sql_instances_history"},
	}
}

// BronzeHistoryGCPSQLInstanceLabel stores historical snapshots of SQL instance labels.
// Links via instance_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPSQLInstanceLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPSQLInstanceLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("instance_history_id").
			Comment("Links to parent BronzeHistoryGCPSQLInstance"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		field.String("key").NotEmpty(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPSQLInstanceLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("instance_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPSQLInstanceLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_sql_instance_labels_history"},
	}
}
