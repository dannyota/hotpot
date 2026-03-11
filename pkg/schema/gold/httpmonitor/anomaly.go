package httpmonitor

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	goldmixin "danny.vn/hotpot/pkg/schema/gold/mixin"
)

// GoldHttpmonitorAnomaly holds detected API traffic anomalies.
type GoldHttpmonitorAnomaly struct {
	ent.Schema
}

func (GoldHttpmonitorAnomaly) Mixin() []ent.Mixin {
	return []ent.Mixin{
		goldmixin.Timestamp{},
	}
}

func (GoldHttpmonitorAnomaly) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("endpoint_id").Optional(),
		field.String("source_id").NotEmpty(),
		field.String("anomaly_type").
			NotEmpty().
			Comment("traffic_spike, traffic_drop, error_burst, 5xx_burst, new_endpoint, scanner_detected, single_ip_flood, method_mismatch, new_user_agent, ua_share_shift, automated_client, ua_spoofing, new_source_ip, geo_shift, external_on_internal, ip_concentration, ip_rotation, new_asn, hosting_provider, asn_concentration, response_size_anomaly, off_hours_spike, endpoint_enumeration, sanctioned_country, path_traversal, sql_injection_probe, command_injection_probe, xss_probe, ssrf_probe, auth_failure_burst, credential_stuffing, otp_brute_force, privilege_escalation_probe, password_reset_flood, registration_abuse, rate_limit_triggered, auth_success_after_burst, bulk_data_extraction, pagination_scraping"),
		field.String("severity").
			NotEmpty().
			Comment("info, low, medium, high, critical"),
		field.Time("window_start"),
		field.Time("window_end"),
		field.String("uri").Optional(),
		field.String("method").Optional(),
		field.Float("baseline_value").Optional(),
		field.Float("actual_value").Optional(),
		field.Float("deviation").
			Optional().
			Comment("Z-score or ratio"),
		field.String("description").Optional(),
		field.JSON("evidence_json", json.RawMessage{}).Optional(),
	}
}

func (GoldHttpmonitorAnomaly) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("anomaly_type"),
		index.Fields("severity"),
		index.Fields("source_id", "window_start"),
		index.Fields("endpoint_id"),
	}
}

func (GoldHttpmonitorAnomaly) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "httpmonitor_anomalies"},
	}
}
