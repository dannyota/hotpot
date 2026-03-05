-- Add new schema named "gold"
CREATE SCHEMA IF NOT EXISTS "gold";
-- Create "lifecycle_software" table
CREATE TABLE "gold"."lifecycle_software" (
  "resource_id" character varying NOT NULL,
  "detected_at" timestamptz NOT NULL,
  "first_detected_at" timestamptz NOT NULL,
  "machine_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "version" character varying NULL,
  "classification" character varying NOT NULL,
  "eol_product_slug" character varying NULL,
  "eol_product_name" character varying NULL,
  "eol_category" character varying NULL,
  "eol_cycle" character varying NULL,
  "eol_date" timestamptz NULL,
  "eoas_date" timestamptz NULL,
  "eol_status" character varying NOT NULL,
  "latest_version" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "goldlifecyclesoftware_classification" to table: "lifecycle_software"
CREATE INDEX "goldlifecyclesoftware_classification" ON "gold"."lifecycle_software" ("classification");
-- Create index "goldlifecyclesoftware_eol_product_slug" to table: "lifecycle_software"
CREATE INDEX "goldlifecyclesoftware_eol_product_slug" ON "gold"."lifecycle_software" ("eol_product_slug");
-- Create index "goldlifecyclesoftware_eol_status" to table: "lifecycle_software"
CREATE INDEX "goldlifecyclesoftware_eol_status" ON "gold"."lifecycle_software" ("eol_status");
-- Create index "goldlifecyclesoftware_machine_id" to table: "lifecycle_software"
CREATE INDEX "goldlifecyclesoftware_machine_id" ON "gold"."lifecycle_software" ("machine_id");
-- Create index "goldlifecyclesoftware_machine_id_name" to table: "lifecycle_software"
CREATE UNIQUE INDEX "goldlifecyclesoftware_machine_id_name" ON "gold"."lifecycle_software" ("machine_id", "name");
