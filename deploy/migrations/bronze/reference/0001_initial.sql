-- Add new schema named "bronze"
CREATE SCHEMA IF NOT EXISTS "bronze";
-- Create "reference_cpe" table
CREATE TABLE "bronze"."reference_cpe" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "part" character varying NOT NULL,
  "cpe_vendor" character varying NOT NULL,
  "cpe_product" character varying NOT NULL,
  "cpe_version" character varying NOT NULL,
  "title" character varying NULL,
  "deprecated" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferencecpe_collected_at" to table: "reference_cpe"
CREATE INDEX "bronzereferencecpe_collected_at" ON "bronze"."reference_cpe" ("collected_at");
-- Create index "bronzereferencecpe_cpe_vendor_cpe_product" to table: "reference_cpe"
CREATE INDEX "bronzereferencecpe_cpe_vendor_cpe_product" ON "bronze"."reference_cpe" ("cpe_vendor", "cpe_product");
-- Create index "bronzereferencecpe_part" to table: "reference_cpe"
CREATE INDEX "bronzereferencecpe_part" ON "bronze"."reference_cpe" ("part");
-- Create "reference_eol_cycles" table
CREATE TABLE "bronze"."reference_eol_cycles" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "product" character varying NOT NULL,
  "cycle" character varying NOT NULL,
  "release_date" timestamptz NULL,
  "eoas" timestamptz NULL,
  "eol" timestamptz NULL,
  "eoes" timestamptz NULL,
  "latest" character varying NULL,
  "latest_release_date" timestamptz NULL,
  "lts" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferenceeolcycle_collected_at" to table: "reference_eol_cycles"
CREATE INDEX "bronzereferenceeolcycle_collected_at" ON "bronze"."reference_eol_cycles" ("collected_at");
-- Create index "bronzereferenceeolcycle_product" to table: "reference_eol_cycles"
CREATE INDEX "bronzereferenceeolcycle_product" ON "bronze"."reference_eol_cycles" ("product");
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
-- Create "reference_eol_products" table
CREATE TABLE "bronze"."reference_eol_products" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "category" character varying NOT NULL,
  "tags" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferenceeolproduct_category" to table: "reference_eol_products"
CREATE INDEX "bronzereferenceeolproduct_category" ON "bronze"."reference_eol_products" ("category");
-- Create index "bronzereferenceeolproduct_collected_at" to table: "reference_eol_products"
CREATE INDEX "bronzereferenceeolproduct_collected_at" ON "bronze"."reference_eol_products" ("collected_at");
-- Create "reference_os_core_rules" table
CREATE TABLE "bronze"."reference_os_core_rules" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "rule_type" character varying NOT NULL,
  "os_type" character varying NULL,
  "value" character varying NOT NULL,
  "description" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferenceoscorerule_collected_at" to table: "reference_os_core_rules"
CREATE INDEX "bronzereferenceoscorerule_collected_at" ON "bronze"."reference_os_core_rules" ("collected_at");
-- Create index "bronzereferenceoscorerule_os_type" to table: "reference_os_core_rules"
CREATE INDEX "bronzereferenceoscorerule_os_type" ON "bronze"."reference_os_core_rules" ("os_type");
-- Create index "bronzereferenceoscorerule_rule_type" to table: "reference_os_core_rules"
CREATE INDEX "bronzereferenceoscorerule_rule_type" ON "bronze"."reference_os_core_rules" ("rule_type");
-- Create "reference_rpm_packages" table
CREATE TABLE "bronze"."reference_rpm_packages" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "package_name" character varying NOT NULL,
  "repo" character varying NOT NULL,
  "arch" character varying NOT NULL,
  "version" character varying NULL,
  "rpm_group" character varying NULL,
  "summary" character varying NULL,
  "url" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferencerpmpackage_collected_at" to table: "reference_rpm_packages"
CREATE INDEX "bronzereferencerpmpackage_collected_at" ON "bronze"."reference_rpm_packages" ("collected_at");
-- Create index "bronzereferencerpmpackage_package_name" to table: "reference_rpm_packages"
CREATE INDEX "bronzereferencerpmpackage_package_name" ON "bronze"."reference_rpm_packages" ("package_name");
-- Create index "bronzereferencerpmpackage_repo" to table: "reference_rpm_packages"
CREATE INDEX "bronzereferencerpmpackage_repo" ON "bronze"."reference_rpm_packages" ("repo");
-- Create "reference_software_match_rules" table
CREATE TABLE "bronze"."reference_software_match_rules" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "product_slug" character varying NOT NULL,
  "rule_type" character varying NOT NULL,
  "os_type" character varying NULL,
  "value" character varying NOT NULL,
  "extra_value" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferencesoftwarematchrule_collected_at" to table: "reference_software_match_rules"
CREATE INDEX "bronzereferencesoftwarematchrule_collected_at" ON "bronze"."reference_software_match_rules" ("collected_at");
-- Create index "bronzereferencesoftwarematchrule_product_slug" to table: "reference_software_match_rules"
CREATE INDEX "bronzereferencesoftwarematchrule_product_slug" ON "bronze"."reference_software_match_rules" ("product_slug");
-- Create index "bronzereferencesoftwarematchrule_rule_type" to table: "reference_software_match_rules"
CREATE INDEX "bronzereferencesoftwarematchrule_rule_type" ON "bronze"."reference_software_match_rules" ("rule_type");
-- Create "reference_ubuntu_packages" table
CREATE TABLE "bronze"."reference_ubuntu_packages" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "package_name" character varying NOT NULL,
  "release" character varying NOT NULL,
  "component" character varying NOT NULL,
  "section" character varying NOT NULL,
  "description" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferenceubuntupackage_collected_at" to table: "reference_ubuntu_packages"
CREATE INDEX "bronzereferenceubuntupackage_collected_at" ON "bronze"."reference_ubuntu_packages" ("collected_at");
-- Create index "bronzereferenceubuntupackage_package_name" to table: "reference_ubuntu_packages"
CREATE INDEX "bronzereferenceubuntupackage_package_name" ON "bronze"."reference_ubuntu_packages" ("package_name");
-- Create index "bronzereferenceubuntupackage_section" to table: "reference_ubuntu_packages"
CREATE INDEX "bronzereferenceubuntupackage_section" ON "bronze"."reference_ubuntu_packages" ("section");
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
-- Create "reference_xeol_products" table
CREATE TABLE "bronze"."reference_xeol_products" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "permalink" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferencexeolproduct_collected_at" to table: "reference_xeol_products"
CREATE INDEX "bronzereferencexeolproduct_collected_at" ON "bronze"."reference_xeol_products" ("collected_at");
-- Create index "bronzereferencexeolproduct_name" to table: "reference_xeol_products"
CREATE INDEX "bronzereferencexeolproduct_name" ON "bronze"."reference_xeol_products" ("name");
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
