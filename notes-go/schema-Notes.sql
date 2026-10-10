CREATE TABLE "Tags" (
  "id" VARCHAR(255) PRIMARY KEY,
  "value" VARCHAR(255) UNIQUE NOT NULL,
  FOREIGN KEY ("id") REFERENCES "notes_on_tags"("tag_id")
);

CREATE TABLE "metadata_note" (
  "id" VARCHAR(255) PRIMARY KEY,
  "item_id" VARCHAR(255) UNIQUE NOT NULL,
  "readonly" BOOLEAN NOT NULL,
  "pinned" BOOLEAN NOT NULL,
  FOREIGN KEY ("item_id") REFERENCES "Items"("id")
);

CREATE TABLE "Items" (
  "id" VARCHAR(255) PRIMARY KEY,
  "parent_id" VARCHAR(255),
  "created_at" VARCHAR(255) NOT NULL,
  "updated_at" DATE,
  "deleted_at" DATE,
  "kind (note | folder)" VARCHAR(255) NOT NULL,
  "value" TEXT
);

CREATE INDEX "idx_Items_kind (note | folder)" ON "Items" ("kind (note | folder)");

CREATE TABLE "notes_on_tags" (
  "tag_id" VARCHAR(255) NOT NULL,
  "note_id" VARCHAR(255) NOT NULL,
  FOREIGN KEY ("note_id") REFERENCES "metadata_note"("id")
);

CREATE TABLE "metadata_folder" (
  "id" VARCHAR(255) PRIMARY KEY,
  "item_id" VARCHAR(255) UNIQUE NOT NULL,
  FOREIGN KEY ("item_id") REFERENCES "Items"("id")
);