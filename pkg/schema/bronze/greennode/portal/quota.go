package portal

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodePortalQuota represents a GreenNode quota in the bronze layer.
type BronzeGreenNodePortalQuota struct {
	ent.Schema
}

func (BronzeGreenNodePortalQuota) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodePortalQuota) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite key: project_id:quota_name"),
		field.String("name").
			NotEmpty().
			Comment("Quota name (e.g. ram, cpu, volume_storage)"),
		field.String("description").
			Optional(),
		field.String("type").
			Optional(),
		field.Int("limit_value").
			Comment("Maximum allowed value"),
		field.Int("used_value").
			Comment("Currently used value"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodePortalQuota) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodePortalQuota) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_portal_quotas"},
	}
}
