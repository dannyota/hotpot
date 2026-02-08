package vpcaccess

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BronzeHistoryGCPVPCAccessConnector stores historical snapshots of GCP VPC Access connectors.
// Uses resource_id for lookup (full resource name), with valid_from/valid_to for time range.
type BronzeHistoryGCPVPCAccessConnector struct {
	ent.Schema
}

func (BronzeHistoryGCPVPCAccessConnector) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze connector by resource_id"),
		field.Time("valid_from").
			Immutable().
			Comment("Start of validity period"),
		field.Time("valid_to").
			Optional().
			Nillable().
			Comment("End of validity period (null = current)"),
		field.Time("collected_at").
			Comment("Timestamp when this snapshot was collected"),

		// All connector fields (same as bronze.BronzeGCPVPCAccessConnector)
		field.String("network").
			Optional(),
		field.String("ip_cidr_range").
			Optional(),
		field.String("state").
			Optional(),
		field.Int32("min_throughput").
			Optional(),
		field.Int32("max_throughput").
			Optional(),
		field.Int32("min_instances").
			Optional(),
		field.Int32("max_instances").
			Optional(),
		field.String("machine_type").
			Optional(),
		field.String("region").
			Optional(),

		// JSONB fields
		field.JSON("subnet_json", json.RawMessage{}).
			Optional(),
		field.JSON("connected_projects_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPVPCAccessConnector) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPVPCAccessConnector) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_vpc_access_connectors_history"},
	}
}
