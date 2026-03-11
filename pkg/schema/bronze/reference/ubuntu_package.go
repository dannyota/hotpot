package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceUbuntuPackage represents an Ubuntu package from the Packages index.
type BronzeReferenceUbuntuPackage struct {
	ent.Schema
}

func (BronzeReferenceUbuntuPackage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceUbuntuPackage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite key: {release}:{component}:{package}"),
		field.String("package_name").
			Comment("Package name (e.g. rsyslog)"),
		field.String("release").
			Comment("Ubuntu release codename (e.g. noble, jammy)"),
		field.String("component").
			Comment("Repository component (main, universe)"),
		field.String("section").
			Comment("Package section (e.g. admin, utils, web, libs)"),
		field.String("description").
			Optional().
			Comment("Package description line"),
	}
}

func (BronzeReferenceUbuntuPackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("package_name"),
		index.Fields("section"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceUbuntuPackage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_ubuntu_packages"},
	}
}
