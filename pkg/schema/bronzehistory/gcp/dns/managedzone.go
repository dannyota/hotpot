package dns

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPDNSManagedZone stores historical snapshots of GCP Cloud DNS managed zones.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPDNSManagedZone struct {
	ent.Schema
}

func (BronzeHistoryGCPDNSManagedZone) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPDNSManagedZone) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze managed zone by resource_id"),

		// All managed zone fields (same as bronze.BronzeGCPDNSManagedZone)
		field.String("name").
			NotEmpty(),
		field.String("dns_name").
			Optional(),
		field.String("description").
			Optional(),
		field.String("visibility").
			Optional(),
		field.String("creation_time").
			Optional(),

		// JSONB fields
		field.JSON("dnssec_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("private_visibility_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("forwarding_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("peering_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("cloud_logging_config_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPDNSManagedZone) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPDNSManagedZone) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_dns_managed_zones_history"},
	}
}

// BronzeHistoryGCPDNSManagedZoneLabel stores historical snapshots of managed zone labels.
// Links via managed_zone_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPDNSManagedZoneLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPDNSManagedZoneLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("managed_zone_history_id").
			Comment("Links to parent BronzeHistoryGCPDNSManagedZone"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		field.String("key").NotEmpty(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPDNSManagedZoneLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("managed_zone_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPDNSManagedZoneLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_dns_managed_zone_labels_history"},
	}
}
