package pki

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeVaultPKICertificate represents a Vault PKI certificate in the bronze layer.
type BronzeVaultPKICertificate struct {
	ent.Schema
}

func (BronzeVaultPKICertificate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeVaultPKICertificate) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("{vault_name}/{mount_path}/{serial}"),
		field.String("vault_name").
			NotEmpty(),
		field.String("mount_path").
			NotEmpty(),
		field.String("serial_number").
			NotEmpty(),
		field.String("common_name").
			Optional(),
		field.String("issuer_cn").
			Optional(),
		field.String("subject_cn").
			Optional(),
		field.String("sans").
			Optional().
			Comment("JSON array of all SANs"),
		field.String("key_type").
			Optional().
			Comment("RSA, ECDSA, etc."),
		field.Int("key_bits").
			Optional(),
		field.String("signing_algo").
			Optional(),
		field.Time("not_before").
			Optional(),
		field.Time("not_after").
			Optional(),
		field.Bool("is_revoked").
			Default(false),
		field.Time("revoked_at").
			Optional().
			Nillable(),
		field.String("certificate_pem").
			Optional().
			Comment("Full PEM text"),
	}
}

func (BronzeVaultPKICertificate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("vault_name"),
		index.Fields("mount_path"),
		index.Fields("common_name"),
		index.Fields("not_after"),
		index.Fields("is_revoked"),
		index.Fields("collected_at"),
	}
}

func (BronzeVaultPKICertificate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "vault_pki_certificates"},
	}
}
