package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDODatabaseBackup stores historical snapshots of DigitalOcean Database Backups.
type BronzeHistoryDODatabaseBackup struct {
	ent.Schema
}

func (BronzeHistoryDODatabaseBackup) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDODatabaseBackup) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze DatabaseBackup by resource_id"),
		field.String("cluster_id").
			NotEmpty(),
		field.Float("size_gigabytes").
			Default(0),
		field.Time("api_created_at").
			Optional().
			Nillable(),
	}
}

func (BronzeHistoryDODatabaseBackup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("cluster_id"),
	}
}

func (BronzeHistoryDODatabaseBackup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_backups_history"},
	}
}
