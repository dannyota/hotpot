package do

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDODatabaseReplica represents a DigitalOcean Database Replica in the bronze layer.
type BronzeDODatabaseReplica struct {
	ent.Schema
}

func (BronzeDODatabaseReplica) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDODatabaseReplica) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite ID: {clusterID}:{replicaName}"),
		field.String("cluster_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("region").
			Optional(),
		field.String("status").
			Optional(),
		field.String("size").
			Optional(),
		field.Uint64("storage_size_mib").
			Default(0),
		field.String("private_network_uuid").
			Optional(),
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
	}
}

func (BronzeDODatabaseReplica) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_id"),
		index.Fields("region"),
		index.Fields("status"),
		index.Fields("collected_at"),
	}
}

func (BronzeDODatabaseReplica) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_replicas"},
	}
}
