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
