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

// BronzeGCPAppEngineService represents a GCP App Engine service in the bronze layer.
// Fields preserve raw API response data from appengine.apps.services.list.
type BronzeGCPAppEngineService struct {
	ent.Schema
}

func (BronzeGCPAppEngineService) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPAppEngineService) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Service resource name (apps/{app}/services/{svc})"),
		field.String("name").
			NotEmpty().
			Comment("Full path to the Service resource in the API"),

		// SplitJSON contains the traffic split configuration (TrafficSplit).
		//
		//	{"shardBy": "IP", "allocations": {"v1": 0.5, "v2": 0.5}}
		field.JSON("split_json", json.RawMessage{}).
			Optional().
			Comment("Traffic split configuration for version routing"),

		// LabelsJSON contains service labels as key/value pairs.
		//
		//	{"env": "prod", "team": "backend"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional().
			Comment("Labels attached to the service"),

		// NetworkSettingsJSON contains ingress settings for the service.
		//
		//	{"ingressTrafficAllowed": "INGRESS_TRAFFIC_ALLOWED_ALL"}
		field.JSON("network_settings_json", json.RawMessage{}).
			Optional().
			Comment("Network/ingress settings for the service"),

		// Collection metadata
		field.String("project_id").
			NotEmpty().
			Comment("Project ID that owns this service"),
	}
}

func (BronzeGCPAppEngineService) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("application", BronzeGCPAppEngineApplication.Type).
			Ref("services").
			Unique().
			Required(),
	}
}

func (BronzeGCPAppEngineService) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPAppEngineService) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_appengine_services"},
	}
}
