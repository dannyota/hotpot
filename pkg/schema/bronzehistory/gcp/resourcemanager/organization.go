package resourcemanager

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPOrganization stores historical snapshots of GCP organizations.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPOrganization struct {
	ent.Schema
}

func (BronzeHistoryGCPOrganization) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPOrganization) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze organization by resource_id"),

		// All organization fields (same as bronze.BronzeGCPOrganization)
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

func (BronzeHistoryGCPOrganization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPOrganization) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_organizations_history"},
	}
}
