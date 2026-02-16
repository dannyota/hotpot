package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeSecurityPolicy represents a GCP Cloud Armor security policy in the bronze layer.
// Fields preserve raw API response data from compute.securityPolicies.list.
type BronzeGCPComputeSecurityPolicy struct {
	ent.Schema
}

func (BronzeGCPComputeSecurityPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeSecurityPolicy) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID, used as primary key for linking"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// Security policy configuration
		field.String("type").
			Optional().
			Comment("CLOUD_ARMOR, CLOUD_ARMOR_EDGE, or CLOUD_ARMOR_NETWORK"),
		field.String("fingerprint").
			Optional(),

		// RulesJSON contains security policy rules.
		//
		//	[{"priority": 2147483647, "action": "deny(403)", "match": {...}}]
		field.JSON("rules_json", json.RawMessage{}).
			Optional(),

		// AssociationsJSON contains backend service associations.
		//
		//	[{"name": "...", "securityPolicyId": "..."}]
		field.JSON("associations_json", json.RawMessage{}).
			Optional(),

		// AdaptiveProtectionConfigJSON contains adaptive protection settings.
		//
		//	{"layer7DdosDefenseConfig": {"enable": true}}
		field.JSON("adaptive_protection_config_json", json.RawMessage{}).
			Optional(),

		// AdvancedOptionsConfigJSON contains advanced options.
		//
		//	{"jsonParsing": "STANDARD", "logLevel": "NORMAL"}
		field.JSON("advanced_options_config_json", json.RawMessage{}).
			Optional(),

		// DdosProtectionConfigJSON contains DDoS protection settings.
		//
		//	{"ddosProtection": "ADVANCED"}
		field.JSON("ddos_protection_config_json", json.RawMessage{}).
			Optional(),

		// RecaptchaOptionsConfigJSON contains reCAPTCHA options.
		//
		//	{"redirectSiteKey": "..."}
		field.JSON("recaptcha_options_config_json", json.RawMessage{}).
			Optional(),

		// LabelsJSON contains user-defined labels.
		//
		//	{"env": "prod", "team": "security"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeSecurityPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeSecurityPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_security_policies"},
	}
}
