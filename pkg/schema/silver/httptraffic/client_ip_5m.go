package httptraffic

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	inventorymixin "danny.vn/hotpot/pkg/schema/silver/inventory/mixin"
)

// SilverHttptrafficClientIp5m stores per-IP traffic data enriched with GeoIP + ASN.
type SilverHttptrafficClientIp5m struct {
	ent.Schema
}

func (SilverHttptrafficClientIp5m) Mixin() []ent.Mixin {
	return []ent.Mixin{
		inventorymixin.Timestamp{},
	}
}

func (SilverHttptrafficClientIp5m) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("endpoint_id").Optional(),
		field.String("source_id").NotEmpty(),
		field.Time("window_start"),
		field.Time("window_end"),
		field.String("uri").NotEmpty(),
		field.String("method").Optional(),
		field.String("client_ip").NotEmpty(),
		field.String("country_code").Optional().
			Comment("From GeoIP: VN, US, etc."),
		field.String("country_name").Optional().
			Comment("From GeoIP: Vietnam, United States, etc."),
		field.Int("asn").Optional().
			Comment("From ASN lookup: 15169, 13335, etc."),
		field.String("org_name").Optional().
			Comment("From ASN lookup: Google LLC, Cloudflare Inc, etc."),
		field.String("as_domain").Optional().
			Comment("Domain of the AS: google.com, cloudflare.com — from IPinfo"),
		field.String("asn_type").Optional().
			Comment("AS type: isp, hosting, business, education — from IPinfo paid"),
		field.Bool("is_internal").Default(false).
			Comment("RFC1918/loopback/link-local"),
		field.Int64("request_count"),
		field.Bool("is_mapped").Default(false),
	}
}

func (SilverHttptrafficClientIp5m) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("endpoint_id", "window_start"),
		index.Fields("client_ip", "window_start"),
		index.Fields("country_code", "window_start"),
		index.Fields("asn", "window_start"),
		index.Fields("is_internal"),
		index.Fields("window_start"),
	}
}

func (SilverHttptrafficClientIp5m) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "httptraffic_client_ip_5m"},
	}
}
