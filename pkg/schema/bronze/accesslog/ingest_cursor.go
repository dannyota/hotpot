package accesslog

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeAccesslogIngestCursor tracks the ingestion watermark per log source config.
type BronzeAccesslogIngestCursor struct {
	ent.Schema
}

func (BronzeAccesslogIngestCursor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeAccesslogIngestCursor) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").
			StorageKey("cursor_id"),
		field.String("name").
			NotEmpty().
			Comment("Config source name, e.g. \"prod-dmz-nginx\""),
		field.String("source_type").
			NotEmpty().
			Comment("Source type, e.g. \"gcplogging\""),
		field.String("source_key").
			NotEmpty().
			Comment("Hash of source-specific config (project_id, log_bucket, etc.)"),
		field.String("role").
			NotEmpty().
			Comment("Source role: \"primary\" or \"enrichment\""),
		field.Time("last_window_end").
			Comment("Watermark: last successfully ingested window end"),
	}
}

func (BronzeAccesslogIngestCursor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "source_type", "source_key").Unique(),
	}
}

func (BronzeAccesslogIngestCursor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "accesslog_ingest_cursors"},
	}
}
