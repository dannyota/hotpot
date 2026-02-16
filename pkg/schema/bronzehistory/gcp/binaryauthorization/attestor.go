package binaryauthorization

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPBinaryAuthorizationAttestor stores historical snapshots of Binary Authorization attestors.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPBinaryAuthorizationAttestor struct {
	ent.Schema
}

func (BronzeHistoryGCPBinaryAuthorizationAttestor) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPBinaryAuthorizationAttestor) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze binary authorization attestor by resource_id"),

		// All attestor fields (same as bronze.BronzeGCPBinaryAuthorizationAttestor)
		field.String("description").
			Optional(),
		field.JSON("user_owned_grafeas_note_json", json.RawMessage{}).
			Optional(),
		field.String("update_time").
			Optional(),
		field.String("etag").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPBinaryAuthorizationAttestor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPBinaryAuthorizationAttestor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_binaryauthorization_attestors_history"},
	}
}
