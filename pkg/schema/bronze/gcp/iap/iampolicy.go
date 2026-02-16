package iap

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPIAPIAMPolicy represents a GCP IAP IAM policy in the bronze layer.
// Fields preserve raw API response data from iap.GetIamPolicy.
type BronzeGCPIAPIAMPolicy struct {
	ent.Schema
}

func (BronzeGCPIAPIAMPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPIAPIAMPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Resource name for which the IAM policy applies"),
		field.String("name").
			NotEmpty().
			Comment("Resource name for which the IAM policy applies"),
		field.String("etag").
			Optional(),
		field.Int("version").
			Optional(),
		field.JSON("bindings_json", json.RawMessage{}).
			Optional().
			Comment("IAM policy bindings as JSON array"),
		field.JSON("audit_configs_json", json.RawMessage{}).
			Optional().
			Comment("IAM audit configs as JSON array"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPIAPIAMPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeGCPIAPIAMPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_iap_iam_policies"},
	}
}
