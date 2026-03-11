package httptraffic

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	inventorymixin "danny.vn/hotpot/pkg/schema/silver/inventory/mixin"
)

// SilverHttptrafficTraffic5m stores traffic data mapped to endpoints in 5-minute windows.
type SilverHttptrafficTraffic5m struct {
	ent.Schema
}

func (SilverHttptrafficTraffic5m) Mixin() []ent.Mixin {
	return []ent.Mixin{
		inventorymixin.Timestamp{},
	}
}

func (SilverHttptrafficTraffic5m) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("endpoint_id").
			Optional().
			Comment("Matched endpoint (nullable for unmapped)"),
		field.String("source_id").NotEmpty(),
		field.Time("window_start"),
		field.Time("window_end"),
		field.String("uri").NotEmpty(),
		field.String("method").Optional(),
		field.Int("status_code"),
		field.Int64("request_count"),
		field.Float("avg_request_time").Optional(),
		field.Float("max_request_time").Optional(),
		field.Int64("total_body_bytes_sent").Default(0),
		field.Int("unique_client_count").Default(0),
		field.String("access_level").
			Optional().
			Comment("Denormalized from endpoint"),
		field.String("service").
			Optional().
			Comment("Denormalized from endpoint"),
		field.Bool("is_mapped").
			Default(false).
			Comment("Whether this row matched a known endpoint"),
		field.Bool("is_method_mismatch").
			Default(false).
			Comment("URI matched but HTTP method not in endpoint's allowed methods"),
		field.Bool("is_scanner_detected").
			Default(false).
			Comment("Scanner UA detected in this window"),
		field.Bool("is_lfi_detected").
			Default(false).
			Comment("Path traversal / LFI pattern detected in URI"),
		field.Bool("is_sqli_detected").
			Default(false).
			Comment("SQL injection pattern detected in URI"),
		field.Bool("is_rce_detected").
			Default(false).
			Comment("Command injection / RCE pattern detected in URI"),
		field.Bool("is_xss_detected").
			Default(false).
			Comment("XSS pattern detected in URI"),
		field.Bool("is_ssrf_detected").
			Default(false).
			Comment("SSRF pattern detected in URI"),
	}
}

func (SilverHttptrafficTraffic5m) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("endpoint_id", "window_start"),
		index.Fields("source_id", "window_start"),
		index.Fields("is_mapped"),
		index.Fields("is_method_mismatch"),
		index.Fields("is_scanner_detected"),
		index.Fields("is_lfi_detected"),
		index.Fields("is_sqli_detected"),
		index.Fields("is_rce_detected"),
		index.Fields("is_xss_detected"),
		index.Fields("is_ssrf_detected"),
	}
}

func (SilverHttptrafficTraffic5m) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "httptraffic_traffic_5m"},
	}
}
