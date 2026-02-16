package dns

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPDNSManagedZone represents a GCP Cloud DNS managed zone in the bronze layer.
// Fields preserve raw API response data from dns.managedZones.list.
type BronzeGCPDNSManagedZone struct {
	ent.Schema
}

func (BronzeGCPDNSManagedZone) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPDNSManagedZone) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields - ID is uint64, stored as string
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID (uint64 as string), used as primary key for linking"),
		field.String("name").
			NotEmpty(),
		field.String("dns_name").
			Optional().
			Comment("The DNS name of this managed zone, e.g. example.com."),
		field.String("description").
			Optional(),
		field.String("visibility").
			Optional().
			Comment("public or private"),
		field.String("creation_time").
			Optional().
			Comment("RFC 3339 date-time when the zone was created"),

		// JSONB configuration fields
		// DnssecConfigJSON contains DNSSEC configuration.
		//
		//	{"state": "on", "defaultKeySpecs": [...], "nonExistence": "nsec3"}
		field.JSON("dnssec_config_json", json.RawMessage{}).
			Optional(),

		// PrivateVisibilityConfigJSON contains networks this zone is visible to (when private).
		//
		//	{"networks": [{"networkUrl": "..."}]}
		field.JSON("private_visibility_config_json", json.RawMessage{}).
			Optional(),

		// ForwardingConfigJSON contains forwarding configuration for this zone.
		//
		//	{"targetNameServers": [{"ipv4Address": "8.8.8.8", "forwardingPath": "default"}]}
		field.JSON("forwarding_config_json", json.RawMessage{}).
			Optional(),

		// PeeringConfigJSON contains DNS peering configuration.
		//
		//	{"targetNetwork": {"networkUrl": "..."}}
		field.JSON("peering_config_json", json.RawMessage{}).
			Optional(),

		// CloudLoggingConfigJSON contains Cloud Logging configuration.
		//
		//	{"enableLogging": true}
		field.JSON("cloud_logging_config_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPDNSManagedZone) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPDNSManagedZoneLabel.Type),
	}
}

func (BronzeGCPDNSManagedZone) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPDNSManagedZone) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_dns_managed_zones"},
	}
}

// BronzeGCPDNSManagedZoneLabel represents a label on a GCP DNS managed zone.
type BronzeGCPDNSManagedZoneLabel struct {
	ent.Schema
}

func (BronzeGCPDNSManagedZoneLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPDNSManagedZoneLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("managed_zone", BronzeGCPDNSManagedZone.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPDNSManagedZoneLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_dns_managed_zone_labels"},
	}
}
