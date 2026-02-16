package ec2

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryAWSEC2Instance stores historical snapshots of AWS EC2 instances.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryAWSEC2Instance struct {
	ent.Schema
}

func (BronzeHistoryAWSEC2Instance) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryAWSEC2Instance) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze instance by resource_id"),

		// All instance fields (same as bronze.BronzeAWSEC2Instance)
		field.String("name").
			Optional(),
		field.String("instance_type").
			Optional(),
		field.String("state").
			Optional(),
		field.String("vpc_id").
			Optional(),
		field.String("subnet_id").
			Optional(),
		field.String("private_ip_address").
			Optional(),
		field.String("public_ip_address").
			Optional(),
		field.String("ami_id").
			Optional(),
		field.String("key_name").
			Optional(),
		field.Time("launch_time").
			Optional().
			Nillable(),
		field.String("platform").
			Optional(),
		field.String("architecture").
			Optional(),
		field.JSON("security_groups_json", json.RawMessage{}).
			Optional(),
		field.String("account_id").
			NotEmpty(),
		field.String("region").
			NotEmpty(),
	}
}

func (BronzeHistoryAWSEC2Instance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("account_id"),
		index.Fields("region"),
	}
}

func (BronzeHistoryAWSEC2Instance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "aws_ec2_instances_history"},
	}
}

// BronzeHistoryAWSEC2InstanceTag stores historical snapshots of instance tags.
type BronzeHistoryAWSEC2InstanceTag struct {
	ent.Schema
}

func (BronzeHistoryAWSEC2InstanceTag) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("instance_history_id").
			Comment("Links to parent BronzeHistoryAWSEC2Instance"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		field.String("key").
			Optional(),
		field.String("value").
			Optional(),
	}
}

func (BronzeHistoryAWSEC2InstanceTag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("instance_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryAWSEC2InstanceTag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "aws_ec2_instance_tags_history"},
	}
}
