package logging

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPLoggingBucket represents a GCP Cloud Logging bucket in the bronze layer.
// Fields preserve raw API response data from logging.projects.locations.buckets.list.
type BronzeGCPLoggingBucket struct {
	ent.Schema
}

func (BronzeGCPLoggingBucket) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPLoggingBucket) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP bucket resource name, used as primary key for linking"),
		field.String("name").
			NotEmpty().
			Comment("The resource name of the bucket"),
		field.String("description").
			Optional().
			Comment("User-supplied description of the bucket"),
		field.Int32("retention_days").
			Default(0).
			Comment("Maximum number of days to retain log entries"),
		field.Bool("locked").
			Default(false).
			Comment("Whether the bucket is locked and cannot be deleted"),
		field.String("lifecycle_state").
			Optional().
			Comment("Lifecycle state of the bucket (ACTIVE, DELETE_REQUESTED)"),
		field.Bool("analytics_enabled").
			Default(false).
			Comment("Whether log analytics is enabled for this bucket"),
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			Optional().
			Comment("Location of the bucket (e.g. global, us-central1)"),

		// CmekSettingsJSON contains CMEK encryption settings.
		//
		//	{"name": "...", "kmsKeyName": "...", "serviceAccountId": "..."}
		field.JSON("cmek_settings_json", json.RawMessage{}).
			Optional(),

		// IndexConfigsJSON contains indexed fields for log entries.
		//
		//	[{"fieldPath": "jsonPayload.request.status", "type": "INDEX_TYPE_STRING"}]
		field.JSON("index_configs_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeGCPLoggingBucket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPLoggingBucket) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_logging_buckets"},
	}
}
