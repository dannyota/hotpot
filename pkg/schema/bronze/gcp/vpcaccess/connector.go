package vpcaccess

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPVPCAccessConnector represents a GCP Serverless VPC Access connector in the bronze layer.
// Fields preserve raw API response data from vpcaccess.connectors.list.
type BronzeGCPVPCAccessConnector struct {
	ent.Schema
}

func (BronzeGCPVPCAccessConnector) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPVPCAccessConnector) Fields() []ent.Field {
	return []ent.Field{
		// ResourceID is the full resource name (projects/{project}/locations/{region}/connectors/{name}).
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Full resource name of the connector"),
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

		// SubnetJSON contains subnet configuration.
		//
		//	{"name": "subnet-name", "projectId": "project-id"}
		field.JSON("subnet_json", json.RawMessage{}).
			Optional(),

		// ConnectedProjectsJSON contains list of projects connected to this connector.
		//
		//	["project-1", "project-2"]
		field.JSON("connected_projects_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPVPCAccessConnector) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPVPCAccessConnector) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_vpc_access_connectors"},
	}
}
