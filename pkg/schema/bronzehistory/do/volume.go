package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDOVolume stores historical snapshots of DigitalOcean Volumes.
type BronzeHistoryDOVolume struct {
	ent.Schema
}

func (BronzeHistoryDOVolume) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDOVolume) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Volume by resource_id"),
		field.String("name").
			NotEmpty(),
		field.String("region").
			Optional(),
		field.Int64("size_gigabytes").
			Default(0),
		field.String("description").
			Optional(),
		field.JSON("droplet_ids_json", []int{}).
			Optional(),
		field.String("filesystem_type").
			Optional(),
		field.String("filesystem_label").
			Optional(),
		field.JSON("tags_json", []string{}).
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
	}
}

func (BronzeHistoryDOVolume) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("region"),
	}
}

func (BronzeHistoryDOVolume) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_volumes_history"},
	}
}
