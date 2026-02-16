package accesscontextmanager

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

// BronzeGCPAccessContextManagerServicePerimeter represents a GCP Access Context Manager service perimeter in the bronze layer.
// Fields preserve raw API response data from accesscontextmanager.ListServicePerimeters.
type BronzeGCPAccessContextManagerServicePerimeter struct {
	ent.Schema
}

func (BronzeGCPAccessContextManagerServicePerimeter) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPAccessContextManagerServicePerimeter) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Service perimeter resource name (e.g., accessPolicies/123/servicePerimeters/myPerimeter)"),
		field.String("title").
			Optional(),
		field.String("description").
			Optional(),
		field.Int("perimeter_type").
			Default(0).
			Comment("Perimeter type (0=REGULAR, 1=BRIDGE)"),
		field.String("etag").
			Optional(),
		field.Bool("use_explicit_dry_run_spec").
			Default(false).
			Comment("Whether to use explicit dry run spec instead of inheriting from status"),

		// StatusJSON contains the perimeter's enforced configuration.
		//
		//	{
		//	  "resources": ["projects/123"],
		//	  "accessLevels": ["accessPolicies/123/accessLevels/myLevel"],
		//	  "restrictedServices": ["storage.googleapis.com"],
		//	  "vpcAccessibleServices": {...},
		//	  "ingressPolicies": [...],
		//	  "egressPolicies": [...]
		//	}
		field.JSON("status_json", json.RawMessage{}).
			Optional().
			Comment("Enforced service perimeter configuration as JSON"),

		// SpecJSON contains the dry-run configuration.
		//
		//	{
		//	  "resources": ["projects/456"],
		//	  "accessLevels": [...],
		//	  "restrictedServices": [...],
		//	  "vpcAccessibleServices": {...},
		//	  "ingressPolicies": [...],
		//	  "egressPolicies": [...]
		//	}
		field.JSON("spec_json", json.RawMessage{}).
			Optional().
			Comment("Dry-run service perimeter configuration as JSON"),
		field.String("access_policy_name").
			NotEmpty().
			Comment("Parent access policy resource name (e.g., accessPolicies/123)"),
		field.String("organization_id").
			NotEmpty().
			Comment("Organization resource name"),
	}
}

func (BronzeGCPAccessContextManagerServicePerimeter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("access_policy", BronzeGCPAccessContextManagerAccessPolicy.Type).
			Ref("service_perimeters").
			Unique().
			Required(),
	}
}

func (BronzeGCPAccessContextManagerServicePerimeter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
		index.Fields("access_policy_name"),
		index.Fields("perimeter_type"),
	}
}

func (BronzeGCPAccessContextManagerServicePerimeter) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_accesscontextmanager_service_perimeters"},
	}
}
