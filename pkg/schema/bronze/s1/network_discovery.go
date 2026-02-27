package s1

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeS1NetworkDiscovery represents a SentinelOne XDR network discovery device in the bronze layer.
type BronzeS1NetworkDiscovery struct {
	ent.Schema
}

func (BronzeS1NetworkDiscovery) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeS1NetworkDiscovery) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("SentinelOne network discovery device ID"),

		// String fields (Optional)
		field.String("name").Optional(),
		field.String("ip_address").Optional(),
		field.String("domain").Optional(),
		field.String("serial_number").Optional(),
		field.String("category").Optional(),
		field.String("sub_category").Optional(),
		field.String("resource_type").Optional(),
		field.String("os").Optional(),
		field.String("os_family").Optional(),
		field.String("os_version").Optional(),
		field.String("os_name_version").Optional(),
		field.String("architecture").Optional(),
		field.String("manufacturer").Optional(),
		field.String("cpu").Optional(),
		field.String("memory_readable").Optional(),
		field.String("network_name").Optional(),
		field.String("asset_status").Optional(),
		field.String("asset_criticality").Optional(),
		field.String("asset_environment").Optional(),
		field.String("infection_status").Optional(),
		field.String("device_review").Optional(),
		field.String("epp_unsupported_unknown").Optional(),
		field.String("asset_contact_email").Optional(),
		field.String("legacy_identity_policy_name").Optional(),
		field.String("previous_os_type").Optional(),
		field.String("previous_os_version").Optional(),
		field.String("previous_device_function").Optional(),
		field.String("detected_from_site").Optional(),
		field.String("s1_account_id").Optional(),
		field.String("s1_account_name").Optional(),
		field.String("s1_site_id").Optional(),
		field.String("s1_site_name").Optional(),
		field.String("s1_group_id").Optional(),
		field.String("s1_group_name").Optional(),
		field.String("s1_scope_id").Optional(),
		field.String("s1_scope_level").Optional(),
		field.String("s1_scope_path").Optional(),
		field.String("s1_onboarded_account_name").Optional(),
		field.String("s1_onboarded_group_name").Optional(),
		field.String("s1_onboarded_site_name").Optional(),
		field.String("s1_onboarded_scope_level").Optional(),
		field.String("s1_onboarded_scope_path").Optional(),

		// Int fields (Optional)
		field.Int("memory").Optional(),
		field.Int("core_count").Optional(),
		field.Int("s1_management_id").Optional(),
		field.Int("s1_scope_type").Optional(),
		field.Int("s1_onboarded_account_id").Optional(),
		field.Int("s1_onboarded_group_id").Optional(),
		field.Int("s1_onboarded_scope_id").Optional(),
		field.Int("s1_onboarded_site_id").Optional(),

		// Bool fields (Default false)
		field.Bool("is_ad_connector").Default(false),
		field.Bool("is_dc_server").Default(false),
		field.Bool("ads_enabled").Default(false),

		// Time fields (Optional/Nillable)
		field.Time("first_seen_dt").Optional().Nillable(),
		field.Time("last_update_dt").Optional().Nillable(),
		field.Time("last_active_dt").Optional().Nillable(),
		field.Time("last_reboot_dt").Optional().Nillable(),
		field.Time("s1_updated_at").Optional().Nillable(),

		// JSON fields (Optional)
		field.JSON("agent_json", json.RawMessage{}).Optional(),
		field.JSON("network_interfaces_json", json.RawMessage{}).Optional(),
		field.JSON("alerts_json", json.RawMessage{}).Optional(),
		field.JSON("alerts_count_json", json.RawMessage{}).Optional(),
		field.JSON("device_review_log_json", json.RawMessage{}).Optional(),
		field.JSON("identity_json", json.RawMessage{}).Optional(),
		field.JSON("notes_json", json.RawMessage{}).Optional(),
		field.JSON("tags_json", json.RawMessage{}).Optional(),
		field.JSON("missing_coverage_json", json.RawMessage{}).Optional(),
		field.JSON("subnets_json", json.RawMessage{}).Optional(),
		field.JSON("surfaces_json", json.RawMessage{}).Optional(),
		field.JSON("network_names_json", json.RawMessage{}).Optional(),
		field.JSON("risk_factors_json", json.RawMessage{}).Optional(),
		field.JSON("active_coverage_json", json.RawMessage{}).Optional(),
		field.JSON("discovery_methods_json", json.RawMessage{}).Optional(),
		field.JSON("hostnames_json", json.RawMessage{}).Optional(),
		field.JSON("internal_ips_json", json.RawMessage{}).Optional(),
		field.JSON("internal_ips_v6_json", json.RawMessage{}).Optional(),
		field.JSON("mac_addresses_json", json.RawMessage{}).Optional(),
		field.JSON("gateway_ips_json", json.RawMessage{}).Optional(),
		field.JSON("gateway_macs_json", json.RawMessage{}).Optional(),
		field.JSON("tcp_ports_json", json.RawMessage{}).Optional(),
		field.JSON("udp_ports_json", json.RawMessage{}).Optional(),
		field.JSON("ranger_tags_json", json.RawMessage{}).Optional(),
		field.JSON("id_secondary_json", json.RawMessage{}).Optional(),
	}
}

func (BronzeS1NetworkDiscovery) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("asset_status"),
		index.Fields("s1_account_id"),
		index.Fields("s1_site_id"),
		index.Fields("infection_status"),
		index.Fields("collected_at"),
		index.Fields("device_review"),
	}
}

func (BronzeS1NetworkDiscovery) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_network_discoveries"},
	}
}
