package binaryauthorization

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPBinaryAuthorizationAttestor represents a GCP Binary Authorization attestor in the bronze layer.
// Fields preserve raw API response data from binaryauthorization.ListAttestors.
type BronzeGCPBinaryAuthorizationAttestor struct {
	ent.Schema
}

func (BronzeGCPBinaryAuthorizationAttestor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPBinaryAuthorizationAttestor) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Attestor resource name (e.g., projects/123/attestors/my-attestor)"),
		field.String("description").
			Optional(),

		// UserOwnedGrafeasNote as JSON containing the note reference and public keys.
		//
		//	{"noteReference": "projects/.../notes/...", "publicKeys": [...], "delegationServiceAccountEmail": "..."}
		field.JSON("user_owned_grafeas_note_json", json.RawMessage{}).
			Optional().
			Comment("User-owned Grafeas note reference and public keys as JSON"),

		field.String("update_time").
			Optional(),
		field.String("etag").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPBinaryAuthorizationAttestor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPBinaryAuthorizationAttestor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_binaryauthorization_attestors"},
	}
}
