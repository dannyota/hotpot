package binaryauthorization

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPBinaryAuthorizationPolicy represents a GCP Binary Authorization policy in the bronze layer.
// Fields preserve raw API response data from binaryauthorization.GetPolicy.
type BronzeGCPBinaryAuthorizationPolicy struct {
	ent.Schema
}

func (BronzeGCPBinaryAuthorizationPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPBinaryAuthorizationPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Policy resource name (e.g., projects/123/policy)"),
		field.String("description").
			Optional(),
		field.Int("global_policy_evaluation_mode").
			Default(0).
			Comment("Global policy evaluation mode (0=UNSPECIFIED, 1=ENABLE, 2=DISABLE)"),

		// DefaultAdmissionRule as JSON containing evaluation_mode, enforcement_mode,
		// and require_attestations_by.
		//
		//	{"evaluationMode": "...", "enforcementMode": "...", "requireAttestationsBy": [...]}
		field.JSON("default_admission_rule_json", json.RawMessage{}).
			Optional().
			Comment("Default admission rule as JSON"),

		// ClusterAdmissionRules as JSON map from cluster name to AdmissionRule.
		//
		//	{"us-east1-a.cluster-1": {"evaluationMode": "...", ...}}
		field.JSON("cluster_admission_rules_json", json.RawMessage{}).
			Optional().
			Comment("Cluster-specific admission rules as JSON map"),

		// KubernetesNamespaceAdmissionRules as JSON map from namespace to AdmissionRule.
		//
		//	{"namespace-1": {"evaluationMode": "...", ...}}
		field.JSON("kube_namespace_admission_rules_json", json.RawMessage{}).
			Optional().
			Comment("Kubernetes namespace admission rules as JSON map"),

		// IstioServiceIdentityAdmissionRules as JSON map from service identity to AdmissionRule.
		//
		//	{"spiffe://...": {"evaluationMode": "...", ...}}
		field.JSON("istio_service_identity_admission_rules_json", json.RawMessage{}).
			Optional().
			Comment("Istio service identity admission rules as JSON map"),

		field.String("update_time").
			Optional(),
		field.String("etag").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPBinaryAuthorizationPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPBinaryAuthorizationPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_binaryauthorization_policies"},
	}
}
