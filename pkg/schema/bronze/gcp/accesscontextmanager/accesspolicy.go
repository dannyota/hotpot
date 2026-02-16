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

// BronzeGCPAccessContextManagerAccessPolicy represents a GCP Access Context Manager access policy in the bronze layer.
// Fields preserve raw API response data from accesscontextmanager.ListAccessPolicies.
type BronzeGCPAccessContextManagerAccessPolicy struct {
	ent.Schema
}

func (BronzeGCPAccessContextManagerAccessPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPAccessContextManagerAccessPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Access policy resource name (e.g., accessPolicies/123)"),
		field.String("parent").
			NotEmpty().
			Comment("Parent organization (e.g., organizations/123)"),
		field.String("title").
			Optional(),
		field.String("etag").
			Optional(),

		// ScopesJSON contains the access policy scopes (folders/projects the policy applies to).
		//
		//	["projects/123", "folders/456"]
		field.JSON("scopes_json", json.RawMessage{}).
			Optional().
			Comment("Access policy scopes as JSON array"),
		field.String("organization_id").
			NotEmpty().
			Comment("Organization resource name"),
	}
}

func (BronzeGCPAccessContextManagerAccessPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("access_levels", BronzeGCPAccessContextManagerAccessLevel.Type),
		edge.To("service_perimeters", BronzeGCPAccessContextManagerServicePerimeter.Type),
	}
}

func (BronzeGCPAccessContextManagerAccessPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
	}
}

func (BronzeGCPAccessContextManagerAccessPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_accesscontextmanager_access_policies"},
	}
}
