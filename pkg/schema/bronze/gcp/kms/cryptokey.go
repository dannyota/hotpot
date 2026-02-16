package kms

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPKMSCryptoKey represents a GCP Cloud KMS crypto key in the bronze layer.
type BronzeGCPKMSCryptoKey struct {
	ent.Schema
}

func (BronzeGCPKMSCryptoKey) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPKMSCryptoKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("KMS crypto key resource name"),
		field.String("name").
			NotEmpty(),
		field.String("purpose").
			Optional(),
		field.String("create_time").
			Optional(),
		field.String("next_rotation_time").
			Optional(),
		field.String("rotation_period").
			Optional(),
		field.String("destroy_scheduled_duration").
			Optional(),
		field.Bool("import_only").
			Default(false),
		field.String("crypto_key_backend").
			Optional(),

		// VersionTemplateJSON contains the crypto key version template.
		//
		//	{"algorithm": "GOOGLE_SYMMETRIC_ENCRYPTION", "protectionLevel": "SOFTWARE"}
		field.JSON("version_template_json", json.RawMessage{}).
			Optional(),

		// PrimaryJSON contains the primary crypto key version.
		//
		//	{"name": "...", "state": "ENABLED", "algorithm": "..."}
		field.JSON("primary_json", json.RawMessage{}).
			Optional(),

		// LabelsJSON contains user-defined labels.
		//
		//	{"env": "prod", "team": "security"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			Optional(),
		field.String("key_ring_name").
			Optional().
			Comment("Parent key ring resource name"),
	}
}

func (BronzeGCPKMSCryptoKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPKMSCryptoKey) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_kms_crypto_keys"},
	}
}
