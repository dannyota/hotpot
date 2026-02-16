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

// BronzeGCPIAPSettings represents GCP Identity-Aware Proxy settings in the bronze layer.
// Fields preserve raw API response data from iap.GetIapSettings.
type BronzeGCPIAPSettings struct {
	ent.Schema
}

func (BronzeGCPIAPSettings) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPIAPSettings) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("IAP settings resource name (e.g., projects/123/iap_web/compute/services/456)"),
		field.String("name").
			NotEmpty().
			Comment("Resource name for the IAP settings"),
		field.JSON("access_settings_json", json.RawMessage{}).
			Optional().
			Comment("Access settings as JSON (CorsSettings, GcipSettings, OAuthSettings)"),
		field.JSON("application_settings_json", json.RawMessage{}).
			Optional().
			Comment("Application settings as JSON (CsmSettings, AccessDeniedPageSettings)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPIAPSettings) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeGCPIAPSettings) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_iap_settings"},
	}
}
