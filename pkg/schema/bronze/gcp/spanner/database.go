package spanner

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

// BronzeGCPSpannerDatabase represents a GCP Cloud Spanner database in the bronze layer.
type BronzeGCPSpannerDatabase struct {
	ent.Schema
}

func (BronzeGCPSpannerDatabase) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPSpannerDatabase) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Spanner database resource name"),
		field.String("name").
			NotEmpty(),
		field.Int("state").
			Optional().
			Comment("Database state enum value"),
		field.String("create_time").
			Optional(),

		// RestoreInfoJSON contains information about the source used to restore the database.
		//
		//	{"sourceType": "BACKUP", "backupInfo": {"backup": "...", "sourceDatabase": "..."}}
		field.JSON("restore_info_json", json.RawMessage{}).
			Optional(),

		// EncryptionConfigJSON contains the encryption configuration for the database.
		//
		//	{"kmsKeyName": "projects/.../locations/.../keyRings/.../cryptoKeys/..."}
		field.JSON("encryption_config_json", json.RawMessage{}).
			Optional(),

		// EncryptionInfoJSON contains the encryption information for the database.
		//
		//	[{"encryptionType": "CUSTOMER_MANAGED_ENCRYPTION", "kmsKeyVersion": "..."}]
		field.JSON("encryption_info_json", json.RawMessage{}).
			Optional(),

		field.String("version_retention_period").
			Optional(),
		field.String("earliest_version_time").
			Optional(),
		field.String("default_leader").
			Optional(),
		field.Int("database_dialect").
			Optional().
			Comment("Database dialect enum value"),
		field.Bool("enable_drop_protection").
			Default(false),
		field.Bool("reconciling").
			Default(false),

		// Collection metadata
		field.String("instance_name").
			Optional().
			Comment("Parent Spanner instance resource name"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPSpannerDatabase) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", BronzeGCPSpannerInstance.Type).
			Ref("databases").
			Unique(),
	}
}

func (BronzeGCPSpannerDatabase) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPSpannerDatabase) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_spanner_databases"},
	}
}
