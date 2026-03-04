-- Modify "reference_eol_products" table
ALTER TABLE "bronze"."reference_eol_products" ADD COLUMN "tags" jsonb NULL;
-- Create "reference_eol_identifiers" table
CREATE TABLE "bronze"."reference_eol_identifiers" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "product" character varying NOT NULL,
  "identifier_type" character varying NOT NULL,
  "value" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferenceeolidentifier_collected_at" to table: "reference_eol_identifiers"
CREATE INDEX "bronzereferenceeolidentifier_collected_at" ON "bronze"."reference_eol_identifiers" ("collected_at");
-- Create index "bronzereferenceeolidentifier_identifier_type" to table: "reference_eol_identifiers"
CREATE INDEX "bronzereferenceeolidentifier_identifier_type" ON "bronze"."reference_eol_identifiers" ("identifier_type");
-- Create index "bronzereferenceeolidentifier_product" to table: "reference_eol_identifiers"
CREATE INDEX "bronzereferenceeolidentifier_product" ON "bronze"."reference_eol_identifiers" ("product");
