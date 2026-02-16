package resourcemanager

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPFolder stores historical snapshots of GCP folders.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPFolder struct {
	ent.Schema
}

func (BronzeHistoryGCPFolder) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPFolder) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze folder by resource_id"),

		// All folder fields (same as bronze.BronzeGCPFolder)
		field.String("name").
			NotEmpty(),
		field.String("display_name").
			Optional(),
		field.String("state").
			Optional(),
		field.String("parent").
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

func (BronzeHistoryGCPFolder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPFolder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_folders_history"},
	}
}

// BronzeHistoryGCPFolderLabel stores historical snapshots of folder labels.
// Links via folder_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPFolderLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPFolderLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("folder_history_id").
			Comment("Links to parent BronzeHistoryGCPFolder"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Label fields
		field.String("key").
			Optional(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPFolderLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("folder_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPFolderLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_folder_labels_history"},
	}
}
