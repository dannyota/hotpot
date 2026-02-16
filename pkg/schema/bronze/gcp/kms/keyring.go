package kms

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPKMSKeyRing represents a GCP Cloud KMS key ring in the bronze layer.
type BronzeGCPKMSKeyRing struct {
	ent.Schema
}

func (BronzeGCPKMSKeyRing) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPKMSKeyRing) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("KMS key ring resource name"),
		field.String("name").
			NotEmpty(),
		field.String("create_time").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			Optional(),
	}
}

func (BronzeGCPKMSKeyRing) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPKMSKeyRing) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_kms_key_rings"},
	}
}
