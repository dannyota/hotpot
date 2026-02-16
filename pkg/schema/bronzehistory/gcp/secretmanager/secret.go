package secretmanager

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPSecretManagerSecret stores historical snapshots of GCP Secret Manager secrets.
type BronzeHistoryGCPSecretManagerSecret struct {
	ent.Schema
}

func (BronzeHistoryGCPSecretManagerSecret) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPSecretManagerSecret) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze secret by resource_id"),

		// All secret fields
		field.String("name").
			NotEmpty(),
		field.String("create_time").
			Optional(),
		field.String("etag").
			Optional(),

		// JSONB fields
		field.JSON("replication_json", json.RawMessage{}).
			Optional(),
		field.JSON("rotation_json", json.RawMessage{}).
			Optional(),
		field.JSON("topics_json", json.RawMessage{}).
			Optional(),
		field.JSON("version_aliases_json", json.RawMessage{}).
			Optional(),
		field.JSON("annotations_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPSecretManagerSecret) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPSecretManagerSecret) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_secretmanager_secrets_history"},
	}
}

// BronzeHistoryGCPSecretManagerSecretLabel stores historical snapshots of secret labels.
type BronzeHistoryGCPSecretManagerSecretLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPSecretManagerSecretLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("secret_history_id").
			Comment("Links to parent BronzeHistoryGCPSecretManagerSecret"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Label fields
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPSecretManagerSecretLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("secret_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPSecretManagerSecretLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_secretmanager_secret_labels_history"},
	}
}
