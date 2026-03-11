package accesslog

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeAccesslogClientIp stores per-IP request counts in 5-minute windows.
type BronzeAccesslogClientIp struct {
	ent.Schema
}

func (BronzeAccesslogClientIp) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeAccesslogClientIp) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment(`Deterministic: "{source_id}:{window_start}:{uri}:{method}:{client_ip}"`),
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
		field.String("client_ip").
			NotEmpty(),
		field.Int64("request_count"),
	}
}

func (BronzeAccesslogClientIp) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("source_id", "window_start"),
		index.Fields("client_ip", "window_start"),
		index.Fields("window_start"),
	}
}

func (BronzeAccesslogClientIp) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "accesslog_client_ips"},
	}
}
