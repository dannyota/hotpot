package ec2

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

// BronzeAWSEC2Instance represents an AWS EC2 instance in the bronze layer.
// Fields preserve raw API response data from ec2.DescribeInstances.
type BronzeAWSEC2Instance struct {
	ent.Schema
}

func (BronzeAWSEC2Instance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeAWSEC2Instance) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("EC2 instance ID (i-xxxx), used as primary key"),
		field.String("name").
			Optional().
			Comment("Value of the Name tag"),
		field.String("instance_type").
			Optional(),
		field.String("state").
			Optional().
			Comment("Instance state: pending, running, stopping, stopped, etc."),
		field.String("vpc_id").
			Optional(),
		field.String("subnet_id").
			Optional(),
		field.String("private_ip_address").
			Optional(),
		field.String("public_ip_address").
			Optional(),
		field.String("ami_id").
			Optional().
			Comment("AMI image ID used to launch the instance"),
		field.String("key_name").
			Optional().
			Comment("SSH key pair name"),
		field.Time("launch_time").
			Optional().
			Nillable(),
		field.String("platform").
			Optional().
			Comment("Platform: windows or empty for Linux"),
		field.String("architecture").
			Optional().
			Comment("x86_64, arm64, etc."),

		// SecurityGroupsJSON stores security group IDs and names.
		//
		//	[{"group_id": "sg-xxx", "group_name": "my-sg"}, ...]
		field.JSON("security_groups_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("account_id").
			NotEmpty(),
		field.String("region").
			NotEmpty(),
	}
}

func (BronzeAWSEC2Instance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("tags", BronzeAWSEC2InstanceTag.Type),
	}
}

func (BronzeAWSEC2Instance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("state"),
		index.Fields("account_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeAWSEC2Instance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "aws_ec2_instances"},
	}
}

// BronzeAWSEC2InstanceTag represents a tag on an AWS EC2 instance.
type BronzeAWSEC2InstanceTag struct {
	ent.Schema
}

func (BronzeAWSEC2InstanceTag) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value").
			Optional(),
	}
}

func (BronzeAWSEC2InstanceTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", BronzeAWSEC2Instance.Type).
			Ref("tags").
			Unique().
			Required(),
	}
}

func (BronzeAWSEC2InstanceTag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "aws_ec2_instance_tags"},
	}
}
