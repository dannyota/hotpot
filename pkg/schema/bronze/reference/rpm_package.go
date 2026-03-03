package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceRPMPackage represents an RPM package from CentOS Stream or EPEL repositories.
type BronzeReferenceRPMPackage struct {
	ent.Schema
}

func (BronzeReferenceRPMPackage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceRPMPackage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite key: {repo}:{name}:{arch}"),
		field.String("package_name").
			Comment("RPM package name"),
		field.String("repo").
			Comment("Repository identifier (rhel9-baseos, rhel9-appstream, epel9, rhel7-os, rhel7-updates, epel7)"),
		field.String("arch").
			Comment("Architecture (e.g. x86_64, noarch)"),
		field.String("version").
			Optional().
			Comment("Package version (e.g. 8.2102.0-117.el9)"),
		field.String("rpm_group").
			Optional().
			Comment("RPM group (usually Unspecified)"),
		field.String("summary").
			Optional().
			Comment("One-line package description"),
		field.String("url").
			Optional().
			Comment("Upstream project URL"),
	}
}

func (BronzeReferenceRPMPackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("package_name"),
		index.Fields("repo"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceRPMPackage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_rpm_packages"},
	}
}
