-- Create "vault_pki_certificates" table
CREATE TABLE "bronze"."vault_pki_certificates" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "vault_name" character varying NOT NULL,
  "mount_path" character varying NOT NULL,
  "serial_number" character varying NOT NULL,
  "common_name" character varying NULL,
  "issuer_cn" character varying NULL,
  "subject_cn" character varying NULL,
  "sans" character varying NULL,
  "key_type" character varying NULL,
  "key_bits" bigint NULL,
  "signing_algo" character varying NULL,
  "not_before" timestamptz NULL,
  "not_after" timestamptz NULL,
  "is_revoked" boolean NOT NULL DEFAULT false,
  "revoked_at" timestamptz NULL,
  "certificate_pem" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzevaultpkicertificate_collected_at" to table: "vault_pki_certificates"
CREATE INDEX "bronzevaultpkicertificate_collected_at" ON "bronze"."vault_pki_certificates" ("collected_at");
-- Create index "bronzevaultpkicertificate_common_name" to table: "vault_pki_certificates"
CREATE INDEX "bronzevaultpkicertificate_common_name" ON "bronze"."vault_pki_certificates" ("common_name");
-- Create index "bronzevaultpkicertificate_is_revoked" to table: "vault_pki_certificates"
CREATE INDEX "bronzevaultpkicertificate_is_revoked" ON "bronze"."vault_pki_certificates" ("is_revoked");
-- Create index "bronzevaultpkicertificate_mount_path" to table: "vault_pki_certificates"
CREATE INDEX "bronzevaultpkicertificate_mount_path" ON "bronze"."vault_pki_certificates" ("mount_path");
-- Create index "bronzevaultpkicertificate_not_after" to table: "vault_pki_certificates"
CREATE INDEX "bronzevaultpkicertificate_not_after" ON "bronze"."vault_pki_certificates" ("not_after");
-- Create index "bronzevaultpkicertificate_vault_name" to table: "vault_pki_certificates"
CREATE INDEX "bronzevaultpkicertificate_vault_name" ON "bronze"."vault_pki_certificates" ("vault_name");
