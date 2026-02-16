package do

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDODatabaseReplica stores historical snapshots of DigitalOcean Database Replicas.
type BronzeHistoryDODatabaseReplica struct {
	ent.Schema
}

func (BronzeHistoryDODatabaseReplica) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDODatabaseReplica) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze DatabaseReplica by resource_id"),
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

func (BronzeHistoryDODatabaseReplica) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("cluster_id"),
		index.Fields("region"),
		index.Fields("status"),
	}
}

func (BronzeHistoryDODatabaseReplica) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_replicas_history"},
	}
}
