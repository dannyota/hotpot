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

// BronzeGCPComputeSslPolicy represents a GCP Compute Engine SSL policy in the bronze layer.
// Fields preserve raw API response data from compute.sslPolicies.list.
type BronzeGCPComputeSslPolicy struct {
	ent.Schema
}

func (BronzeGCPComputeSslPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeSslPolicy) Fields() []ent.Field {
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

		// SSL policy configuration
		field.String("profile").
			Optional().
			Comment("COMPATIBLE, MODERN, RESTRICTED, or CUSTOM"),
		field.String("min_tls_version").
			Optional().
			Comment("TLS_1_0, TLS_1_1, or TLS_1_2"),
		field.String("fingerprint").
			Optional(),

		// CustomFeaturesJSON contains custom SSL features when profile is CUSTOM.
		//
		//	["TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", ...]
		field.JSON("custom_features_json", json.RawMessage{}).
			Optional(),

		// EnabledFeaturesJSON contains all enabled SSL features.
		//
		//	["TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", ...]
		field.JSON("enabled_features_json", json.RawMessage{}).
			Optional(),

		// WarningsJSON contains warnings about the SSL policy configuration.
		//
		//	[{"code": "...", "message": "...", "data": [...]}]
		field.JSON("warnings_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeSslPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeSslPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_ssl_policies"},
	}
}
