package dataproc

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPDataprocCluster represents a GCP Dataproc cluster in the bronze layer.
// Fields preserve raw API response data from dataproc.projects.regions.clusters.list.
type BronzeGCPDataprocCluster struct {
	ent.Schema
}

func (BronzeGCPDataprocCluster) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPDataprocCluster) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Cluster resource name (projects/{project}/regions/{region}/clusters/{cluster})"),
		field.String("cluster_name").
			NotEmpty(),
		field.String("cluster_uuid").
			Optional(),

		// ConfigJSON contains the ClusterConfig covering master/worker instance groups,
		// software configuration, networking, encryption, autoscaling, and more.
		//
		//	{"masterConfig": {...}, "workerConfig": {...}, "softwareConfig": {...}, ...}
		field.JSON("config_json", json.RawMessage{}).
			Optional(),

		// StatusJSON contains the current ClusterStatus.
		//
		//	{"state": "RUNNING", "detail": "...", "stateStartTime": "..."}
		field.JSON("status_json", json.RawMessage{}).
			Optional(),

		// StatusHistoryJSON contains previous ClusterStatus entries.
		//
		//	[{"state": "CREATING", "stateStartTime": "..."}, ...]
		field.JSON("status_history_json", json.RawMessage{}).
			Optional(),

		// LabelsJSON contains user-provided labels.
		//
		//	{"key1": "value1", "key2": "value2"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// MetricsJSON contains cluster daemon metrics (HDFS, YARN stats).
		//
		//	{"hdfsMetrics": {...}, "yarnMetrics": {...}}
		field.JSON("metrics_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			NotEmpty().
			Comment("Region where the cluster runs"),
	}
}

func (BronzeGCPDataprocCluster) Edges() []ent.Edge {
	return nil
}

func (BronzeGCPDataprocCluster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("location"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPDataprocCluster) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_dataproc_clusters"},
	}
}
