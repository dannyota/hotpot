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
-- Create "reference_eol_products" table
CREATE TABLE "bronze"."reference_eol_products" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "category" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferenceeolproduct_category" to table: "reference_eol_products"
CREATE INDEX "bronzereferenceeolproduct_category" ON "bronze"."reference_eol_products" ("category");
-- Create index "bronzereferenceeolproduct_collected_at" to table: "reference_eol_products"
CREATE INDEX "bronzereferenceeolproduct_collected_at" ON "bronze"."reference_eol_products" ("collected_at");
-- Create "reference_xeol_products" table
CREATE TABLE "bronze"."reference_xeol_products" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "purl" character varying NULL,
  "permalink" character varying NULL,
  "eol" timestamptz NULL,
  "eol_bool" boolean NOT NULL DEFAULT false,
  "latest_cycle" character varying NULL,
  "release_date" timestamptz NULL,
  "latest" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferencexeolproduct_collected_at" to table: "reference_xeol_products"
CREATE INDEX "bronzereferencexeolproduct_collected_at" ON "bronze"."reference_xeol_products" ("collected_at");
-- Create index "bronzereferencexeolproduct_name" to table: "reference_xeol_products"
CREATE INDEX "bronzereferencexeolproduct_name" ON "bronze"."reference_xeol_products" ("name");
-- Create index "bronzereferencexeolproduct_purl" to table: "reference_xeol_products"
CREATE INDEX "bronzereferencexeolproduct_purl" ON "bronze"."reference_xeol_products" ("purl");
