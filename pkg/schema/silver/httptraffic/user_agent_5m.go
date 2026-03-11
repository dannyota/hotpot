package httptraffic

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	inventorymixin "danny.vn/hotpot/pkg/schema/silver/inventory/mixin"
)

// SilverHttptrafficUserAgent5m stores per-UA traffic data enriched with endpoint info.
type SilverHttptrafficUserAgent5m struct {
	ent.Schema
}

func (SilverHttptrafficUserAgent5m) Mixin() []ent.Mixin {
	return []ent.Mixin{
		inventorymixin.Timestamp{},
	}
}

func (SilverHttptrafficUserAgent5m) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("endpoint_id").Optional(),
		field.String("source_id").NotEmpty(),
		field.Time("window_start"),
		field.Time("window_end"),
		field.String("uri").NotEmpty(),
		field.String("method").Optional(),
		field.String("user_agent").NotEmpty().
			Comment("Full UA string"),
		field.String("ua_family").Optional().
			Comment("Normalized family: chrome, curl, python-requests, etc."),
		field.Int64("request_count"),
		field.Bool("is_mapped").Default(false),
	}
}

func (SilverHttptrafficUserAgent5m) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("endpoint_id", "window_start"),
		index.Fields("ua_family", "window_start"),
		index.Fields("window_start"),
	}
}

func (SilverHttptrafficUserAgent5m) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "httptraffic_user_agent_5m"},
	}
}
