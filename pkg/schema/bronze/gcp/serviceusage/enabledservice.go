package serviceusage

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPServiceUsageEnabledService represents a GCP enabled service in the bronze layer.
type BronzeGCPServiceUsageEnabledService struct {
	ent.Schema
}

func (BronzeGCPServiceUsageEnabledService) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPServiceUsageEnabledService) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Service resource name (projects/{project}/services/{service})"),
		field.String("name").
			NotEmpty().
			Comment("Service resource name"),
		field.String("parent").
			NotEmpty().
			Comment("Parent resource (projects/{project})"),

		// ConfigJSON contains the service configuration (title, APIs, documentation, quota, etc.).
		//
		//	{"name": "compute.googleapis.com", "title": "Compute Engine API", "apis": [...], ...}
		field.JSON("config_json", json.RawMessage{}).
			Optional().
			Comment("ServiceConfig as JSON"),

		field.Int("state").
			Default(0).
			Comment("Service state: 0=STATE_UNSPECIFIED, 1=DISABLED, 2=ENABLED"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPServiceUsageEnabledService) Edges() []ent.Edge {
	return nil
}

func (BronzeGCPServiceUsageEnabledService) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPServiceUsageEnabledService) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_serviceusage_enabled_services"},
	}
}
