package run

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

// BronzeGCPRunService represents a GCP Cloud Run service in the bronze layer.
// Fields preserve raw API response data from run.services.list.
type BronzeGCPRunService struct {
	ent.Schema
}

func (BronzeGCPRunService) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPRunService) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Service resource name (projects/{project}/locations/{location}/services/{service})"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("uid").
			Optional(),
		field.Int64("generation").
			Optional(),

		// LabelsJSON contains user-provided labels.
		//
		//	{"key1": "value1", "key2": "value2"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// AnnotationsJSON contains user-provided annotations.
		//
		//	{"key1": "value1", "key2": "value2"}
		field.JSON("annotations_json", json.RawMessage{}).
			Optional(),

		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),
		field.String("delete_time").
			Optional(),
		field.String("creator").
			Optional(),
		field.String("last_modifier").
			Optional(),
		field.Int("ingress").
			Optional().
			Comment("IngressTraffic enum value"),
		field.Int("launch_stage").
			Optional().
			Comment("LaunchStage enum value"),

		// TemplateJSON contains the RevisionTemplate configuration.
		//
		//	{"containers": [...], "scaling": {...}, "serviceAccount": "...", ...}
		field.JSON("template_json", json.RawMessage{}).
			Optional(),

		// TrafficJSON contains traffic allocation targets.
		//
		//	[{"type": "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST", "percent": 100}]
		field.JSON("traffic_json", json.RawMessage{}).
			Optional(),

		field.String("uri").
			Optional(),
		field.Int64("observed_generation").
			Optional(),

		// TerminalConditionJSON contains the readiness status of the service.
		//
		//	{"type": "Ready", "state": "CONDITION_SUCCEEDED", ...}
		field.JSON("terminal_condition_json", json.RawMessage{}).
			Optional(),

		// ConditionsJSON contains diagnostics for sub-resources.
		//
		//	[{"type": "...", "state": "...", "message": "..."}]
		field.JSON("conditions_json", json.RawMessage{}).
			Optional(),

		field.String("latest_ready_revision").
			Optional(),
		field.String("latest_created_revision").
			Optional(),

		// TrafficStatusesJSON contains detailed status for traffic targets.
		//
		//	[{"type": "...", "revision": "...", "percent": 100, "uri": "..."}]
		field.JSON("traffic_statuses_json", json.RawMessage{}).
			Optional(),

		field.Bool("reconciling").
			Default(false),
		field.String("etag").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			NotEmpty(),
	}
}

func (BronzeGCPRunService) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("revisions", BronzeGCPRunRevision.Type),
	}
}

func (BronzeGCPRunService) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("location"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPRunService) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_run_services"},
	}
}
