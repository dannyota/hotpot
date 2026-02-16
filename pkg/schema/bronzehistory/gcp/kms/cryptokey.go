package kms

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPKMSCryptoKey stores historical snapshots of GCP KMS crypto keys.
type BronzeHistoryGCPKMSCryptoKey struct {
	ent.Schema
}

func (BronzeHistoryGCPKMSCryptoKey) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPKMSCryptoKey) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze crypto key by resource_id"),

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

		// JSONB fields
		field.JSON("version_template_json", json.RawMessage{}).
			Optional(),
		field.JSON("primary_json", json.RawMessage{}).
			Optional(),
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			Optional(),
		field.String("key_ring_name").
			Optional(),
	}
}

func (BronzeHistoryGCPKMSCryptoKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPKMSCryptoKey) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_kms_crypto_keys_history"},
	}
}
