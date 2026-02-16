package resourcemanager

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPOrganization represents a GCP organization in the bronze layer.
// Fields preserve raw API response data from cloudresourcemanager.organizations.search.
type BronzeGCPOrganization struct {
	ent.Schema
}

func (BronzeGCPOrganization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPOrganization) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Organization resource name (e.g., organizations/123)"),
		field.String("name").
			NotEmpty(),
		field.String("display_name").
			Optional(),
		field.String("state").
			Optional(),
		field.String("directory_customer_id").
			Optional(),
		field.String("etag").
			Optional(),
		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),
		field.String("delete_time").
			Optional(),
	}
}

func (BronzeGCPOrganization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("state"),
	}
}

func (BronzeGCPOrganization) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_organizations"},
	}
}
