-- Modify "reference_xeol_products" table
ALTER TABLE "bronze"."reference_xeol_products" DROP COLUMN "purl", DROP COLUMN "eol", DROP COLUMN "eol_bool", DROP COLUMN "latest_cycle", DROP COLUMN "release_date", DROP COLUMN "latest";
-- Create "reference_xeol_cycles" table
CREATE TABLE "bronze"."reference_xeol_cycles" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "product_id" character varying NOT NULL,
  "release_cycle" character varying NOT NULL,
  "eol" timestamptz NULL,
  "eol_bool" boolean NOT NULL DEFAULT false,
  "latest_release" character varying NULL,
  "latest_release_date" timestamptz NULL,
  "release_date" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferencexeolcycle_collected_at" to table: "reference_xeol_cycles"
CREATE INDEX "bronzereferencexeolcycle_collected_at" ON "bronze"."reference_xeol_cycles" ("collected_at");
-- Create index "bronzereferencexeolcycle_product_id" to table: "reference_xeol_cycles"
CREATE INDEX "bronzereferencexeolcycle_product_id" ON "bronze"."reference_xeol_cycles" ("product_id");
-- Create "reference_xeol_purls" table
CREATE TABLE "bronze"."reference_xeol_purls" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "product_id" character varying NOT NULL,
  "purl" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferencexeolpurl_collected_at" to table: "reference_xeol_purls"
CREATE INDEX "bronzereferencexeolpurl_collected_at" ON "bronze"."reference_xeol_purls" ("collected_at");
-- Create index "bronzereferencexeolpurl_product_id" to table: "reference_xeol_purls"
CREATE INDEX "bronzereferencexeolpurl_product_id" ON "bronze"."reference_xeol_purls" ("product_id");
-- Create index "bronzereferencexeolpurl_purl" to table: "reference_xeol_purls"
CREATE INDEX "bronzereferencexeolpurl_purl" ON "bronze"."reference_xeol_purls" ("purl");
-- Create "reference_xeol_vulns" table
CREATE TABLE "bronze"."reference_xeol_vulns" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "product_id" character varying NOT NULL,
  "version" character varying NOT NULL,
  "issue_count" bigint NOT NULL,
  "issues" text NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferencexeolvuln_collected_at" to table: "reference_xeol_vulns"
CREATE INDEX "bronzereferencexeolvuln_collected_at" ON "bronze"."reference_xeol_vulns" ("collected_at");
-- Create index "bronzereferencexeolvuln_product_id" to table: "reference_xeol_vulns"
CREATE INDEX "bronzereferencexeolvuln_product_id" ON "bronze"."reference_xeol_vulns" ("product_id");
