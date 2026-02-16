package secretmanager

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

// BronzeGCPSecretManagerSecret represents a GCP Secret Manager secret in the bronze layer.
type BronzeGCPSecretManagerSecret struct {
	ent.Schema
}

func (BronzeGCPSecretManagerSecret) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPSecretManagerSecret) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Secret resource name"),
		field.String("name").
			NotEmpty(),
		field.String("create_time").
			Optional(),
		field.String("etag").
			Optional(),

		// ReplicationJSON contains the replication policy configuration.
		//
		//	{"automatic": {}} or {"userManaged": {"replicas": [...]}}
		field.JSON("replication_json", json.RawMessage{}).
			Optional(),

		// RotationJSON contains the rotation policy configuration.
		//
		//	{"nextRotationTime": "...", "rotationPeriod": "..."}
		field.JSON("rotation_json", json.RawMessage{}).
			Optional(),

		// TopicsJSON contains Pub/Sub topic configurations for notifications.
		//
		//	[{"name": "projects/.../topics/..."}]
		field.JSON("topics_json", json.RawMessage{}).
			Optional(),

		// VersionAliasesJSON contains version alias mappings.
		//
		//	{"alias1": 1, "alias2": 2}
		field.JSON("version_aliases_json", json.RawMessage{}).
			Optional(),

		// AnnotationsJSON contains user-managed annotations.
		//
		//	{"key1": "value1", "key2": "value2"}
		field.JSON("annotations_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPSecretManagerSecret) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPSecretManagerSecretLabel.Type),
	}
}

func (BronzeGCPSecretManagerSecret) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPSecretManagerSecret) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_secretmanager_secrets"},
	}
}

// BronzeGCPSecretManagerSecretLabel represents a label attached to a GCP Secret Manager secret.
type BronzeGCPSecretManagerSecretLabel struct {
	ent.Schema
}

func (BronzeGCPSecretManagerSecretLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPSecretManagerSecretLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("secret", BronzeGCPSecretManagerSecret.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPSecretManagerSecretLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_secretmanager_secret_labels"},
	}
}
