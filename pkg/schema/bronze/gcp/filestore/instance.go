package filestore

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPFilestoreInstance represents a GCP Filestore instance in the bronze layer.
type BronzeGCPFilestoreInstance struct {
	ent.Schema
}

func (BronzeGCPFilestoreInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPFilestoreInstance) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Instance resource name (projects/*/locations/*/instances/*)"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Int("state").
			Default(0).
			Comment("Instance state (0=UNSPECIFIED, 1=CREATING, 2=READY, 3=REPAIRING, 4=DELETING, 6=ERROR, 7=RESTORING, 8=SUSPENDED, 9=SUSPENDING, 10=RESUMING, 11=REVERTING, 12=PROMOTING)"),
		field.String("status_message").
			Optional(),
		field.String("create_time").
			Optional(),
		field.Int("tier").
			Default(0).
			Comment("Instance tier (0=UNSPECIFIED, 1=STANDARD, 2=PREMIUM, 3=BASIC_HDD, 4=BASIC_SSD, 5=HIGH_SCALE_SSD, 6=ENTERPRISE, 7=ZONAL, 8=REGIONAL)"),

		// LabelsJSON contains user-defined labels.
		//
		//	{"env": "prod", "team": "backend"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// FileSharesJSON contains file share configurations.
		//
		//	[{"name": "vol1", "capacityGb": "1024", "nfsExportOptions": [...]}]
		field.JSON("file_shares_json", json.RawMessage{}).
			Optional(),

		// NetworksJSON contains VPC network configurations.
		//
		//	[{"network": "default", "modes": ["MODE_IPV4"], "reservedIpRange": "...", "ipAddresses": [...], "connectMode": "..."}]
		field.JSON("networks_json", json.RawMessage{}).
			Optional(),

		field.String("etag").
			Optional(),
		field.Bool("satisfies_pzs").
			Default(false),
		field.Bool("satisfies_pzi").
			Default(false),
		field.String("kms_key_name").
			Optional(),

		// SuspensionReasonsJSON contains reasons the instance is suspended.
		//
		//	[0, 1]
		field.JSON("suspension_reasons_json", json.RawMessage{}).
			Optional(),

		field.Int64("max_capacity_gb").
			Optional().
			Comment("Output-only maximum capacity of the instance"),
		field.Int("protocol").
			Default(0).
			Comment("File access protocol (0=UNSPECIFIED, 1=NFS_V3, 2=NFS_V4_1)"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			Optional(),
	}
}

func (BronzeGCPFilestoreInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPFilestoreInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_filestore_instances"},
	}
}
