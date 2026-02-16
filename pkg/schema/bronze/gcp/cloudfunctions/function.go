package cloudfunctions

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPCloudFunctionsFunction represents a GCP Cloud Function in the bronze layer.
type BronzeGCPCloudFunctionsFunction struct {
	ent.Schema
}

func (BronzeGCPCloudFunctionsFunction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPCloudFunctionsFunction) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Function resource name (projects/*/locations/*/functions/*)"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Int("environment").
			Default(0).
			Comment("Function environment (0=UNSPECIFIED, 1=GEN_1, 2=GEN_2)"),
		field.Int("state").
			Default(0).
			Comment("Function state (0=UNSPECIFIED, 1=ACTIVE, 2=FAILED, 3=DEPLOYING, 4=DELETING, 5=UNKNOWN)"),

		// BuildConfigJSON contains the build configuration.
		//
		//	{"runtime": "go121", "entryPoint": "Handler", "source": {...}, ...}
		field.JSON("build_config_json", json.RawMessage{}).
			Optional(),

		// ServiceConfigJSON contains the service deployment configuration.
		//
		//	{"service": "...", "availableMemory": "256M", "environmentVariables": {...}, "vpcConnector": "...", ...}
		field.JSON("service_config_json", json.RawMessage{}).
			Optional(),

		// EventTriggerJSON contains the Eventarc trigger configuration.
		//
		//	{"trigger": "...", "triggerRegion": "...", "eventType": "...", "eventFilters": [...]}
		field.JSON("event_trigger_json", json.RawMessage{}).
			Optional(),

		// StateMessagesJSON contains output-only state messages.
		//
		//	[{"severity": "ERROR", "type": "...", "message": "..."}]
		field.JSON("state_messages_json", json.RawMessage{}).
			Optional(),

		field.String("update_time").
			Optional(),
		field.String("create_time").
			Optional(),

		// LabelsJSON contains user-defined labels.
		//
		//	{"env": "prod", "team": "backend"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		field.String("kms_key_name").
			Optional(),
		field.String("url").
			Optional().
			Comment("Output-only deployed URL for the function"),
		field.Bool("satisfies_pzs").
			Default(false),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			Optional(),
	}
}

func (BronzeGCPCloudFunctionsFunction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPCloudFunctionsFunction) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_cloudfunctions_functions"},
	}
}
