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

// BronzeGCPAccessContextManagerAccessLevel represents a GCP Access Context Manager access level in the bronze layer.
// Fields preserve raw API response data from accesscontextmanager.ListAccessLevels.
type BronzeGCPAccessContextManagerAccessLevel struct {
	ent.Schema
}

func (BronzeGCPAccessContextManagerAccessLevel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPAccessContextManagerAccessLevel) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Access level resource name (e.g., accessPolicies/123/accessLevels/myLevel)"),
		field.String("title").
			Optional(),
		field.String("description").
			Optional(),

		// BasicJSON contains the basic access level conditions.
		//
		//	{
		//	  "conditions": [...],
		//	  "combiningFunction": "AND" | "OR"
		//	}
		field.JSON("basic_json", json.RawMessage{}).
			Optional().
			Comment("Basic access level conditions as JSON"),

		// CustomJSON contains the custom access level CEL expression.
		//
		//	{
		//	  "expr": {
		//	    "expression": "...",
		//	    "title": "...",
		//	    "description": "..."
		//	  }
		//	}
		field.JSON("custom_json", json.RawMessage{}).
			Optional().
			Comment("Custom access level CEL expression as JSON"),
		field.String("access_policy_name").
			NotEmpty().
			Comment("Parent access policy resource name (e.g., accessPolicies/123)"),
		field.String("organization_id").
			NotEmpty().
			Comment("Organization resource name"),
	}
}

func (BronzeGCPAccessContextManagerAccessLevel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("access_policy", BronzeGCPAccessContextManagerAccessPolicy.Type).
			Ref("access_levels").
			Unique().
			Required(),
	}
}

func (BronzeGCPAccessContextManagerAccessLevel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
		index.Fields("access_policy_name"),
	}
}

func (BronzeGCPAccessContextManagerAccessLevel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_accesscontextmanager_access_levels"},
	}
}
