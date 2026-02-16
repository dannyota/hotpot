package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDODatabasePool stores historical snapshots of DigitalOcean Database Connection Pools.
type BronzeHistoryDODatabasePool struct {
	ent.Schema
}

func (BronzeHistoryDODatabasePool) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDODatabasePool) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze DatabasePool by resource_id"),
		field.String("cluster_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("user").
			Optional(),
		field.Int("size").
			Default(0),
		field.String("database").
			Optional(),
		field.String("mode").
			Optional(),
	}
}

func (BronzeHistoryDODatabasePool) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("cluster_id"),
		index.Fields("mode"),
	}
}

func (BronzeHistoryDODatabasePool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_pools_history"},
	}
}
