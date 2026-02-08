package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeForwardingRule represents a GCP Compute Engine regional forwarding rule in the bronze layer.
// Fields preserve raw API response data from compute.forwardingRules.aggregatedList.
type BronzeGCPComputeForwardingRule struct {
	ent.Schema
}

func (BronzeGCPComputeForwardingRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeForwardingRule) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID, used as primary key for linking"),
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

		// JSONB fields (primitive arrays and nested objects)
		field.JSON("ports_json", []interface{}{}).
			Optional().
			Comment("Port numbers (array)"),
		field.JSON("source_ip_ranges_json", []interface{}{}).
			Optional().
			Comment("Source IP ranges (array)"),
		field.JSON("metadata_filters_json", []interface{}{}).
			Optional().
			Comment("Metadata filters (array of objects)"),
		field.JSON("service_directory_registrations_json", []interface{}{}).
			Optional().
			Comment("Service Directory registrations (array of objects)"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeForwardingRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPComputeForwardingRuleLabel.Type),
	}
}

func (BronzeGCPComputeForwardingRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeForwardingRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_forwarding_rules"},
	}
}

// BronzeGCPComputeForwardingRuleLabel represents a label attached to a GCP Compute forwarding rule.
type BronzeGCPComputeForwardingRuleLabel struct {
	ent.Schema
}

func (BronzeGCPComputeForwardingRuleLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPComputeForwardingRuleLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("forwarding_rule", BronzeGCPComputeForwardingRule.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeForwardingRuleLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_forwarding_rule_labels"},
	}
}

// BronzeGCPComputeGlobalForwardingRule represents a GCP Compute Engine global forwarding rule in the bronze layer.
// Fields preserve raw API response data from compute.globalForwardingRules.list.
type BronzeGCPComputeGlobalForwardingRule struct {
	ent.Schema
}

func (BronzeGCPComputeGlobalForwardingRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeGlobalForwardingRule) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID, used as primary key for linking"),
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

		// JSONB fields (primitive arrays and nested objects)
		field.JSON("ports_json", []interface{}{}).
			Optional().
			Comment("Port numbers (array)"),
		field.JSON("source_ip_ranges_json", []interface{}{}).
			Optional().
			Comment("Source IP ranges (array)"),
		field.JSON("metadata_filters_json", []interface{}{}).
			Optional().
			Comment("Metadata filters (array of objects)"),
		field.JSON("service_directory_registrations_json", []interface{}{}).
			Optional().
			Comment("Service Directory registrations (array of objects)"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeGlobalForwardingRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPComputeGlobalForwardingRuleLabel.Type),
	}
}

func (BronzeGCPComputeGlobalForwardingRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeGlobalForwardingRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_global_forwarding_rules"},
	}
}

// BronzeGCPComputeGlobalForwardingRuleLabel represents a label attached to a GCP Compute global forwarding rule.
type BronzeGCPComputeGlobalForwardingRuleLabel struct {
	ent.Schema
}

func (BronzeGCPComputeGlobalForwardingRuleLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPComputeGlobalForwardingRuleLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("global_forwarding_rule", BronzeGCPComputeGlobalForwardingRule.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeGlobalForwardingRuleLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_global_forwarding_rule_labels"},
	}
}
