package accesscontextmanager

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPAccessContextManagerAccessPolicy stores historical snapshots of access policies.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPAccessContextManagerAccessPolicy struct {
	ent.Schema
}

func (BronzeHistoryGCPAccessContextManagerAccessPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPAccessContextManagerAccessPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze access policy by resource_id"),

		// All access policy fields (same as bronze.BronzeGCPAccessContextManagerAccessPolicy)
		field.String("parent").
			NotEmpty(),
		field.String("title").
			Optional(),
		field.String("etag").
			Optional(),
		field.JSON("scopes_json", json.RawMessage{}).
			Optional(),
		field.String("organization_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPAccessContextManagerAccessPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPAccessContextManagerAccessPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_accesscontextmanager_access_policies_history"},
	}
}
