package pki

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryVaultPKICertificate stores historical snapshots of Vault PKI certificates.
type BronzeHistoryVaultPKICertificate struct {
	ent.Schema
}

func (BronzeHistoryVaultPKICertificate) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryVaultPKICertificate) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze certificate by resource_id"),
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
			Optional(),
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
			Optional(),
	}
}

func (BronzeHistoryVaultPKICertificate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("vault_name"),
	}
}

func (BronzeHistoryVaultPKICertificate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "vault_pki_certificates_history"},
	}
}
