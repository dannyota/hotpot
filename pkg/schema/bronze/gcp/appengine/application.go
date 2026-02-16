package appengine

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

// BronzeGCPAppEngineApplication represents a GCP App Engine application in the bronze layer.
// Fields preserve raw API response data from appengine.apps.get.
type BronzeGCPAppEngineApplication struct {
	ent.Schema
}

func (BronzeGCPAppEngineApplication) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPAppEngineApplication) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Application resource name (apps/{app})"),
		field.String("name").
			NotEmpty().
			Comment("Full path to the Application resource in the API"),
		field.String("auth_domain").
			Optional().
			Comment("Google Apps authentication domain"),
		field.String("location_id").
			Optional().
			Comment("Location from which this application runs"),
		field.String("code_bucket").
			Optional().
			Comment("GCS bucket for storing files associated with this application"),
		field.String("default_cookie_expiration").
			Optional().
			Comment("Cookie expiration policy as a duration string"),
		field.Int32("serving_status").
			Optional().
			Comment("Serving status enum: UNSPECIFIED=0, SERVING=1, USER_DISABLED=2, SYSTEM_DISABLED=3"),
		field.String("default_hostname").
			Optional().
			Comment("Hostname used to reach this application"),
		field.String("default_bucket").
			Optional().
			Comment("GCS bucket for storing content"),
		field.String("gcr_domain").
			Optional().
			Comment("Google Container Registry domain for managed build docker images"),
		field.Int32("database_type").
			Optional().
			Comment("Cloud Firestore/Datastore database type enum: UNSPECIFIED=0, CLOUD_DATASTORE=1, CLOUD_FIRESTORE=2, CLOUD_DATASTORE_COMPATIBILITY=3"),

		// FeatureSettingsJSON contains the feature-specific settings.
		//
		//	{"splitHealthChecks": true, "useContainerOptimizedOs": false}
		field.JSON("feature_settings_json", json.RawMessage{}).
			Optional().
			Comment("Feature-specific settings for the application"),

		// IapJSON contains Identity-Aware Proxy configuration.
		//
		//	{"enabled": true, "oauth2ClientId": "...", "oauth2ClientSecretSha256": "..."}
		field.JSON("iap_json", json.RawMessage{}).
			Optional().
			Comment("Identity-Aware Proxy configuration"),

		// DispatchRulesJSON contains HTTP path dispatch rules.
		//
		//	[{"domain": "*", "path": "/...", "service": "default"}]
		field.JSON("dispatch_rules_json", json.RawMessage{}).
			Optional().
			Comment("HTTP path dispatch rules for requests"),

		// Collection metadata
		field.String("project_id").
			NotEmpty().
			Comment("Project ID that owns this application"),
	}
}

func (BronzeGCPAppEngineApplication) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("services", BronzeGCPAppEngineService.Type),
	}
}

func (BronzeGCPAppEngineApplication) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPAppEngineApplication) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_appengine_applications"},
	}
}
