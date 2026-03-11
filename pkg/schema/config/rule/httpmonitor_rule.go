package rule

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ConfigHttpmonitorRule defines a detection rule for the HTTP monitor pipeline.
// All 46 rules from the catalog are seeded — both live and planned — so the
// admin UI shows the full rule inventory.
type ConfigHttpmonitorRule struct {
	ent.Schema
}

func (ConfigHttpmonitorRule) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("rule_id"),
		field.String("rule_key").Unique().Immutable().
			Comment("Code-facing identifier, e.g. traffic_spike_high, geo_shift_new_country"),
		field.String("anomaly_type").NotEmpty().
			Comment("Type written to anomalies: traffic_spike, 5xx_burst, scanner_detected, ..."),
		field.String("severity").NotEmpty().
			Comment("critical, high, medium, low, info"),
		field.String("category").NotEmpty().
			Comment("rate, error, suspicious, endpoint, ua, ip_geo, asn, auth, data, partner, injection, evasion"),
		field.String("name").NotEmpty().
			Comment("Human-readable rule name"),
		field.String("description").NotEmpty().
			Comment("What triggers this rule"),
		field.String("baseline").Optional().
			Comment("What the rule compares against, e.g. '24h rolling avg', '7-day UA baseline'"),
		field.JSON("thresholds_json", map[string]float64{}).Optional().
			Comment("Threshold parameters, e.g. {\"z_score\": 3, \"min_pct\": 0.1}"),
		field.String("status").Default("live").
			Comment("live, planned, deprecated"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (ConfigHttpmonitorRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("anomaly_type"),
		index.Fields("category"),
		index.Fields("status"),
	}
}

func (ConfigHttpmonitorRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "httpmonitor_rules"},
	}
}
