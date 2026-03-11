package rule

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ConfigHostingIndicator stores domains and keywords used to detect hosting/cloud
// providers from IP-to-ASN data. Manageable via admin UI.
//
// Indicator types:
//   - domain:  exact AS domain match (e.g. "amazon.com", "akamai.com")
//   - keyword: substring match in AS domain (e.g. "host", "cloud", "vps")
type ConfigHostingIndicator struct {
	ent.Schema
}

func (ConfigHostingIndicator) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("indicator_type").Values("domain", "keyword").
			Comment("domain = exact match, keyword = substring match"),
		field.String("value").NotEmpty().
			Comment("The domain or keyword to match against AS domain"),
		field.String("description").Optional().
			Comment("Human-readable note, e.g. 'AWS', 'Akamai CDN'"),
		field.String("country").Optional().
			Comment("ISO 3166-1 alpha-2 code (VN, US, JP), NULL for global"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (ConfigHostingIndicator) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("indicator_type", "value").Unique(),
		index.Fields("indicator_type", "is_active"),
		index.Fields("country"),
	}
}

func (ConfigHostingIndicator) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "hosting_indicators"},
	}
}
