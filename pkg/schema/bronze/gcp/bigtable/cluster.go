package bigtable

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

// BronzeGCPBigtableCluster represents a GCP Bigtable cluster in the bronze layer.
// Fields preserve raw API response data from bigtable.admin.v2.clusters.list.
type BronzeGCPBigtableCluster struct {
	ent.Schema
}

func (BronzeGCPBigtableCluster) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPBigtableCluster) Fields() []ent.Field {
	return []ent.Field{
		// Resource name: projects/{project}/instances/{instance}/clusters/{cluster}
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Cluster resource name"),
		field.String("location").
			Optional().
			Comment("Zone or region location of the cluster"),
		field.Int32("state").
			Optional().
			Comment("Cluster state: 0=STATE_NOT_KNOWN, 1=READY, 2=CREATING, 3=RESIZING, 4=DISABLED"),
		field.Int32("serve_nodes").
			Optional().
			Comment("Number of nodes allocated to the cluster"),
		field.Int32("default_storage_type").
			Optional().
			Comment("Default storage type: 0=STORAGE_TYPE_UNSPECIFIED, 1=SSD, 2=HDD"),

		// EncryptionConfigJSON contains CMEK encryption configuration.
		//
		//	{"kmsKeyName": "projects/.../cryptoKeys/..."}
		field.JSON("encryption_config_json", json.RawMessage{}).
			Optional(),

		// ClusterConfigJSON contains autoscaling configuration (ClusterAutoscalingConfig).
		//
		//	{"autoscalingConfig": {"autoscalingLimits": {...}, "autoscalingTargets": {...}}}
		field.JSON("cluster_config_json", json.RawMessage{}).
			Optional(),

		// Parent instance reference
		field.String("instance_name").
			NotEmpty().
			Comment("Parent instance resource name"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPBigtableCluster) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", BronzeGCPBigtableInstance.Type).
			Ref("clusters").
			Unique().
			Required(),
	}
}

func (BronzeGCPBigtableCluster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
		index.Fields("instance_name"),
	}
}

func (BronzeGCPBigtableCluster) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_bigtable_clusters"},
	}
}
