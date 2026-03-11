package accesslog

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeAccesslogHttpCount stores aggregated HTTP traffic counts per 5-min window.
// Append-only: stores ALL URIs (matched + unmatched) for 30-day retention.
type BronzeAccesslogHttpCount struct {
	ent.Schema
}

func (BronzeAccesslogHttpCount) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeAccesslogHttpCount) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Deterministic: \"{source_id}:{window_start}:{uri}:{method}:{status}\""),
		field.String("source_id").
			NotEmpty(),
		field.Time("window_start").
			Immutable(),
		field.Time("window_end").
			Immutable(),
		field.String("http_host").
			Optional(),
		field.String("uri").
			NotEmpty(),
		field.String("method").
			Optional(),
		field.Int("status_code"),
		field.Int64("request_count"),
		field.Float("total_request_time").
			Optional().
			Comment("Sum of request times, for computing avg later"),
		field.Float("max_request_time").
			Optional(),
		field.Int64("total_body_bytes_sent").
			Default(0),
	}
}

func (BronzeAccesslogHttpCount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("source_id", "window_start"),
		index.Fields("uri", "window_start"),
		index.Fields("window_start"),
	}
}

func (BronzeAccesslogHttpCount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "accesslog_http_counts"},
	}
}
