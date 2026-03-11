package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceXeolVuln represents a vulnerability entry from the xeol EOL database.
type BronzeReferenceXeolVuln struct {
	ent.Schema
}

func (BronzeReferenceXeolVuln) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceXeolVuln) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite key: {product_id}:{version}"),
		field.String("product_id").
			Comment("Reference to xeol product ID"),
		field.String("version").
			Comment("Affected version string"),
		field.Int("issue_count").
			Comment("Number of vulnerability issues"),
		field.Text("issues").
			Comment("Vulnerability IDs (e.g. GHSA-xxx, CVE-xxx)"),
	}
}

func (BronzeReferenceXeolVuln) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceXeolVuln) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_xeol_vulns"},
	}
}
