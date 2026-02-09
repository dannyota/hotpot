package resourcemanager

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPProject stores historical snapshots of GCP projects.
// Uses project_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPProject struct {
	ent.Schema
}

func (BronzeHistoryGCPProject) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPProject) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("project_id").
			NotEmpty().
			Comment("Link to bronze project by project_id"),

		// All project fields (same as bronze.BronzeGCPProject)
		field.String("project_number").
			NotEmpty(),
		field.String("display_name").
			Optional(),
		field.String("state").
			Optional(),
		field.String("parent").
			Optional(),
		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),
		field.String("delete_time").
			Optional(),
		field.String("etag").
			Optional(),
	}
}

func (BronzeHistoryGCPProject) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPProject) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_projects_history"},
	}
}

// BronzeHistoryGCPProjectLabel stores historical snapshots of project labels.
// Links via project_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPProjectLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPProjectLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("project_history_id").
			Comment("Links to parent BronzeHistoryGCPProject"),
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

func (BronzeHistoryGCPProjectLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPProjectLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_project_labels_history"},
	}
}
