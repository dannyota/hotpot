package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceCPE represents a CPE entry from the NVD CPE Dictionary.
type BronzeReferenceCPE struct {
	ent.Schema
}

func (BronzeReferenceCPE) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceCPE) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("CPE name string (e.g. cpe:2.3:a:rsyslog:rsyslog:8.2001.0:*:*:*:*:*:*:*)"),
		field.String("part").
			Comment("a (application) or o (operating system)"),
		field.String("cpe_vendor").
			Comment("Vendor name from CPE"),
		field.String("cpe_product").
			Comment("Product name from CPE"),
		field.String("cpe_version").
			Comment("Version from CPE"),
		field.String("title").
			Optional().
			Comment("English title from NVD"),
		field.Bool("deprecated").
			Default(false).
			Comment("NVD deprecated flag"),
	}
}

func (BronzeReferenceCPE) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("part"),
		index.Fields("cpe_vendor", "cpe_product"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceCPE) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_cpe"},
	}
}
