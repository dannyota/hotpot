package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BronzeHistoryGCPComputeForwardingRule stores historical snapshots of GCP Compute regional forwarding rules.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeForwardingRule struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeForwardingRule) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze forwarding rule by resource_id"),
		field.Time("valid_from").
			Immutable().
			Comment("Start of validity period"),
		field.Time("valid_to").
			Optional().
			Nillable().
			Comment("End of validity period (null = current)"),
		field.Time("collected_at").
			Comment("Timestamp when this snapshot was collected"),

		// All forwarding rule fields (same as bronze.BronzeGCPComputeForwardingRule)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("ip_address").
			Optional(),
		field.String("ip_protocol").
			Optional(),
		field.Bool("all_ports").
			Default(false),
		field.Bool("allow_global_access").
			Default(false),
		field.Bool("allow_psc_global_access").
			Default(false),
		field.String("backend_service").
			Optional(),
		field.String("base_forwarding_rule").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("external_managed_backend_bucket_migration_state").
			Optional(),
		field.Float32("external_managed_backend_bucket_migration_testing_percentage").
			Optional(),
		field.String("fingerprint").
			Optional(),
		field.String("ip_collection").
			Optional(),
		field.String("ip_version").
			Optional(),
		field.Bool("is_mirroring_collector").
			Default(false),
		field.String("label_fingerprint").
			Optional(),
		field.String("load_balancing_scheme").
			Optional(),
		field.String("network").
			Optional(),
		field.String("network_tier").
			Optional(),
		field.Bool("no_automate_dns_zone").
			Default(false),
		field.String("port_range").
			Optional(),
		field.String("psc_connection_id").
			Optional(),
		field.String("psc_connection_status").
			Optional(),
		field.String("region").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("self_link_with_id").
			Optional(),
		field.String("service_label").
			Optional(),
		field.String("service_name").
			Optional(),
		field.String("subnetwork").
			Optional(),
		field.String("target").
			Optional(),

		// JSONB fields
		field.JSON("ports_json", []interface{}{}).
			Optional(),
		field.JSON("source_ip_ranges_json", []interface{}{}).
			Optional(),
		field.JSON("metadata_filters_json", []interface{}{}).
			Optional(),
		field.JSON("service_directory_registrations_json", []interface{}{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeForwardingRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeForwardingRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_forwarding_rules_history"},
	}
}

// BronzeHistoryGCPComputeForwardingRuleLabel stores historical snapshots of forwarding rule labels.
// Links via forwarding_rule_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeForwardingRuleLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeForwardingRuleLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("forwarding_rule_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeForwardingRule"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Label fields
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPComputeForwardingRuleLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("forwarding_rule_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeForwardingRuleLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_forwarding_rule_labels_history"},
	}
}

// BronzeHistoryGCPComputeGlobalForwardingRule stores historical snapshots of GCP Compute global forwarding rules.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeGlobalForwardingRule struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeGlobalForwardingRule) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze global forwarding rule by resource_id"),
		field.Time("valid_from").
			Immutable().
			Comment("Start of validity period"),
		field.Time("valid_to").
			Optional().
			Nillable().
			Comment("End of validity period (null = current)"),
		field.Time("collected_at").
			Comment("Timestamp when this snapshot was collected"),

		// All global forwarding rule fields (same as bronze.BronzeGCPComputeGlobalForwardingRule)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("ip_address").
			Optional(),
		field.String("ip_protocol").
			Optional(),
		field.Bool("all_ports").
			Default(false),
		field.Bool("allow_global_access").
			Default(false),
		field.Bool("allow_psc_global_access").
			Default(false),
		field.String("backend_service").
			Optional(),
		field.String("base_forwarding_rule").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("external_managed_backend_bucket_migration_state").
			Optional(),
		field.Float32("external_managed_backend_bucket_migration_testing_percentage").
			Optional(),
		field.String("fingerprint").
			Optional(),
		field.String("ip_collection").
			Optional(),
		field.String("ip_version").
			Optional(),
		field.Bool("is_mirroring_collector").
			Default(false),
		field.String("label_fingerprint").
			Optional(),
		field.String("load_balancing_scheme").
			Optional(),
		field.String("network").
			Optional(),
		field.String("network_tier").
			Optional(),
		field.Bool("no_automate_dns_zone").
			Default(false),
		field.String("port_range").
			Optional(),
		field.String("psc_connection_id").
			Optional(),
		field.String("psc_connection_status").
			Optional(),
		field.String("region").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("self_link_with_id").
			Optional(),
		field.String("service_label").
			Optional(),
		field.String("service_name").
			Optional(),
		field.String("subnetwork").
			Optional(),
		field.String("target").
			Optional(),

		// JSONB fields
		field.JSON("ports_json", []interface{}{}).
			Optional(),
		field.JSON("source_ip_ranges_json", []interface{}{}).
			Optional(),
		field.JSON("metadata_filters_json", []interface{}{}).
			Optional(),
		field.JSON("service_directory_registrations_json", []interface{}{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeGlobalForwardingRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeGlobalForwardingRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_global_forwarding_rules_history"},
	}
}

// BronzeHistoryGCPComputeGlobalForwardingRuleLabel stores historical snapshots of global forwarding rule labels.
// Links via global_forwarding_rule_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeGlobalForwardingRuleLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeGlobalForwardingRuleLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("global_forwarding_rule_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeGlobalForwardingRule"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Label fields
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPComputeGlobalForwardingRuleLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("global_forwarding_rule_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeGlobalForwardingRuleLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_global_forwarding_rule_labels_history"},
	}
}
