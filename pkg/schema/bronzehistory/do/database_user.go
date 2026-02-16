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

// BronzeHistoryDODatabaseUser stores historical snapshots of DigitalOcean Database Users.
type BronzeHistoryDODatabaseUser struct {
	ent.Schema
}

func (BronzeHistoryDODatabaseUser) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDODatabaseUser) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze DatabaseUser by resource_id"),
		field.String("cluster_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("role").
			Optional(),
		field.JSON("mysql_settings_json", json.RawMessage{}).
			Optional(),
		field.JSON("settings_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryDODatabaseUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("cluster_id"),
		index.Fields("role"),
	}
}

func (BronzeHistoryDODatabaseUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_users_history"},
	}
}
