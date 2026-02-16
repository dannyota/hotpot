package compute

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

// BronzeGCPComputeFirewall represents a GCP Compute Engine firewall rule in the bronze layer.
// Fields preserve raw API response data from compute.firewalls.list.
type BronzeGCPComputeFirewall struct {
	ent.Schema
}

func (BronzeGCPComputeFirewall) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeFirewall) Fields() []ent.Field {
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
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// Firewall configuration
		field.String("network").
			Optional().
			Comment("Network URL this firewall applies to"),
		field.Int32("priority").
			Default(1000),
		field.String("direction").
			Optional().
			Comment("INGRESS or EGRESS"),
		field.Bool("disabled").
			Default(false),

		// SourceRangesJSON contains CIDR ranges for ingress traffic.
		//
		//	["10.0.0.0/8", "192.168.0.0/16"]
		field.JSON("source_ranges_json", json.RawMessage{}).
			Optional(),

		// DestinationRangesJSON contains CIDR ranges for egress traffic.
		//
		//	["0.0.0.0/0"]
		field.JSON("destination_ranges_json", json.RawMessage{}).
			Optional(),

		// SourceTagsJSON contains source instance tags.
		//
		//	["web-server", "app-server"]
		field.JSON("source_tags_json", json.RawMessage{}).
			Optional(),

		// TargetTagsJSON contains target instance tags.
		//
		//	["db-server"]
		field.JSON("target_tags_json", json.RawMessage{}).
			Optional(),

		// SourceServiceAccountsJSON contains source service accounts.
		//
		//	["sa@project.iam.gserviceaccount.com"]
		field.JSON("source_service_accounts_json", json.RawMessage{}).
			Optional(),

		// TargetServiceAccountsJSON contains target service accounts.
		//
		//	["sa@project.iam.gserviceaccount.com"]
		field.JSON("target_service_accounts_json", json.RawMessage{}).
			Optional(),

		// LogConfigJSON contains firewall logging configuration.
		//
		//	{"enable": true, "metadata": "INCLUDE_ALL_METADATA"}
		field.JSON("log_config_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeFirewall) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("allowed", BronzeGCPComputeFirewallAllowed.Type),
		edge.To("denied", BronzeGCPComputeFirewallDenied.Type),
	}
}

func (BronzeGCPComputeFirewall) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeFirewall) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_firewalls"},
	}
}

// BronzeGCPComputeFirewallAllowed represents an allowed rule in a firewall.
// Data from firewall.allowed[].
type BronzeGCPComputeFirewallAllowed struct {
	ent.Schema
}

func (BronzeGCPComputeFirewallAllowed) Fields() []ent.Field {
	return []ent.Field{
		field.String("ip_protocol").
			NotEmpty().
			Comment("Protocol: tcp, udp, icmp, esp, ah, sctp, ipip, all"),

		// PortsJSON contains port ranges.
		//
		//	["80", "443", "8000-9000"]
		field.JSON("ports_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeGCPComputeFirewallAllowed) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("firewall_ref", BronzeGCPComputeFirewall.Type).
			Ref("allowed").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeFirewallAllowed) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_firewall_alloweds"},
	}
}

// BronzeGCPComputeFirewallDenied represents a denied rule in a firewall.
// Data from firewall.denied[].
type BronzeGCPComputeFirewallDenied struct {
	ent.Schema
}

func (BronzeGCPComputeFirewallDenied) Fields() []ent.Field {
	return []ent.Field{
		field.String("ip_protocol").
			NotEmpty().
			Comment("Protocol: tcp, udp, icmp, esp, ah, sctp, ipip, all"),

		// PortsJSON contains port ranges.
		//
		//	["80", "443", "8000-9000"]
		field.JSON("ports_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeGCPComputeFirewallDenied) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("firewall_ref", BronzeGCPComputeFirewall.Type).
			Ref("denied").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeFirewallDenied) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_firewall_denieds"},
	}
}
