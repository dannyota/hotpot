package binaryauthorization

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPBinaryAuthorizationPolicy stores historical snapshots of Binary Authorization policies.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPBinaryAuthorizationPolicy struct {
	ent.Schema
}

func (BronzeHistoryGCPBinaryAuthorizationPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPBinaryAuthorizationPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze binary authorization policy by resource_id"),

		// All policy fields (same as bronze.BronzeGCPBinaryAuthorizationPolicy)
		field.String("description").
			Optional(),
		field.Int("global_policy_evaluation_mode").
			Default(0),
		field.JSON("default_admission_rule_json", json.RawMessage{}).
			Optional(),
		field.JSON("cluster_admission_rules_json", json.RawMessage{}).
			Optional(),
		field.JSON("kube_namespace_admission_rules_json", json.RawMessage{}).
			Optional(),
		field.JSON("istio_service_identity_admission_rules_json", json.RawMessage{}).
			Optional(),
		field.String("update_time").
			Optional(),
		field.String("etag").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPBinaryAuthorizationPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPBinaryAuthorizationPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_binaryauthorization_policies_history"},
	}
}
