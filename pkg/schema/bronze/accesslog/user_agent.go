package accesslog

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeAccesslogUserAgent stores per-UA request counts in 5-minute windows.
type BronzeAccesslogUserAgent struct {
	ent.Schema
}

func (BronzeAccesslogUserAgent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeAccesslogUserAgent) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment(`Deterministic: "{source_id}:{window_start}:{uri}:{method}:{ua_hash}"`),
		field.String("source_id").
			NotEmpty(),
		field.Time("window_start").
			Immutable(),
		field.Time("window_end").
			Immutable(),
		field.String("uri").
			NotEmpty(),
		field.String("method").
			Default(""),
		field.String("user_agent").
			NotEmpty().
			Comment("Full UA string"),
		field.Int64("request_count"),
	}
}

func (BronzeAccesslogUserAgent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("source_id", "window_start"),
		index.Fields("uri", "window_start"),
		index.Fields("window_start"),
	}
}

func (BronzeAccesslogUserAgent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "accesslog_user_agents"},
	}
}
