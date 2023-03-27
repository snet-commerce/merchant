-- Create "merchants" table
CREATE TABLE "merchants" ("id" uuid NOT NULL, "created_at" timestamp NOT NULL, "updated_at" timestamp NOT NULL, "name" character varying NOT NULL, "description" text NULL, "primary_email" character varying NOT NULL, "secondary_email" character varying NULL, "primary_number" character varying NOT NULL, "secondary_number" character varying NULL, "tenant" uuid NOT NULL, "active" boolean NOT NULL DEFAULT true, PRIMARY KEY ("id"));
-- Create index "merchants_tenant_key" to table: "merchants"
CREATE UNIQUE INDEX "merchants_tenant_key" ON "merchants" ("tenant");
