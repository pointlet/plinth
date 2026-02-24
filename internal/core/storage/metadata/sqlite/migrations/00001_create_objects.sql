-- +goose Up
CREATE TABLE IF NOT EXISTS objects (
  namespace TEXT NOT NULL,
  path TEXT NOT NULL,
  size INTEGER NOT NULL,
  etag TEXT NOT NULL,
  content_type TEXT NOT NULL,
  filename TEXT NOT NULL,
  version TEXT NOT NULL,
  deleted INTEGER NOT NULL DEFAULT 0,
  attributes_json TEXT NOT NULL DEFAULT '{}',
  created_at INTEGER NOT NULL,
  modified_at INTEGER NOT NULL,
  PRIMARY KEY (namespace, path)
);

CREATE INDEX IF NOT EXISTS idx_objects_namespace_path ON objects(namespace, path);
CREATE INDEX IF NOT EXISTS idx_objects_namespace_modified_at ON objects(namespace, modified_at);

-- +goose Down
DROP INDEX IF EXISTS idx_objects_namespace_modified_at;
DROP INDEX IF EXISTS idx_objects_namespace_path;
DROP TABLE IF EXISTS objects;
