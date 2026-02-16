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

// BronzeDODatabaseUser represents a DigitalOcean Database User in the bronze layer.
type BronzeDODatabaseUser struct {
	ent.Schema
}

func (BronzeDODatabaseUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDODatabaseUser) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite ID: {clusterID}:{userName}"),
		field.String("cluster_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("role").
			Optional(),
		field.JSON("mysql_settings_json", json.RawMessage{}).
			Optional().
			Comment("MySQL auth plugin settings (not password)"),
		field.JSON("settings_json", json.RawMessage{}).
			Optional().
			Comment("ACL and other settings"),
	}
}

func (BronzeDODatabaseUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_id"),
		index.Fields("role"),
		index.Fields("collected_at"),
	}
}

func (BronzeDODatabaseUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_users"},
	}
}
