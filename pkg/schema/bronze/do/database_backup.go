package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDODatabaseBackup represents a DigitalOcean Database Backup in the bronze layer.
type BronzeDODatabaseBackup struct {
	ent.Schema
}

func (BronzeDODatabaseBackup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDODatabaseBackup) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite ID: {clusterID}:{createdAt RFC3339}"),
		field.String("cluster_id").
			NotEmpty(),
		field.Float("size_gigabytes").
			Default(0),
		field.Time("api_created_at").
			Optional().
			Nillable(),
	}
}

func (BronzeDODatabaseBackup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeDODatabaseBackup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_backups"},
	}
}
