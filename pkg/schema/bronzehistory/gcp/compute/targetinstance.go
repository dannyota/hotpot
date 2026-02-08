package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeTargetInstance stores historical snapshots of GCP Compute target instances.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeTargetInstance struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeTargetInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeTargetInstance) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze target instance by resource_id"),

		// All target instance fields (same as bronze.BronzeGCPComputeTargetInstance)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("zone").
			Optional(),
		field.String("instance").
			Optional(),
		field.String("network").
			Optional(),
		field.String("nat_policy").
			Optional(),
		field.String("security_policy").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeTargetInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeTargetInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_target_instances_history"},
	}
}
