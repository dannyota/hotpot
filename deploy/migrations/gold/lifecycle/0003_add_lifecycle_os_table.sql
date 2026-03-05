-- Create "lifecycle_os" table
CREATE TABLE "gold"."lifecycle_os" (
  "resource_id" character varying NOT NULL,
  "detected_at" timestamptz NOT NULL,
  "first_detected_at" timestamptz NOT NULL,
  "machine_id" character varying NOT NULL,
  "hostname" character varying NULL,
  "os_type" character varying NULL,
  "os_name" character varying NULL,
  "eol_product_slug" character varying NULL,
  "eol_product_name" character varying NULL,
  "eol_cycle" character varying NULL,
  "eol_date" timestamptz NULL,
  "eoas_date" timestamptz NULL,
  "eoes_date" timestamptz NULL,
  "eol_status" character varying NOT NULL,
  "latest_version" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "goldlifecycleos_machine_id" to table: "lifecycle_os"
CREATE UNIQUE INDEX "goldlifecycleos_machine_id" ON "gold"."lifecycle_os" ("machine_id");
-- Create index "goldlifecycleos_eol_status" to table: "lifecycle_os"
CREATE INDEX "goldlifecycleos_eol_status" ON "gold"."lifecycle_os" ("eol_status");
-- Create index "goldlifecycleos_eol_product_slug" to table: "lifecycle_os"
CREATE INDEX "goldlifecycleos_eol_product_slug" ON "gold"."lifecycle_os" ("eol_product_slug");
