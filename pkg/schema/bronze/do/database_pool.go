package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDODatabasePool represents a DigitalOcean Database Connection Pool in the bronze layer.
type BronzeDODatabasePool struct {
	ent.Schema
}

func (BronzeDODatabasePool) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDODatabasePool) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite ID: {clusterID}:{poolName}"),
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

func (BronzeDODatabasePool) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_id"),
		index.Fields("mode"),
		index.Fields("collected_at"),
	}
}

func (BronzeDODatabasePool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_pools"},
	}
}
