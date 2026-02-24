-- +goose Up
CREATE TABLE IF NOT EXISTS uploads (
  upload_id TEXT PRIMARY KEY,
  namespace TEXT NOT NULL,
  path TEXT NOT NULL,
  filename TEXT NOT NULL,
  content_type TEXT NOT NULL,
  attributes_json TEXT NOT NULL DEFAULT '{}',
  temp_path TEXT NOT NULL,
  bytes_written INTEGER NOT NULL DEFAULT 0,
  expected_size INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_uploads_status_updated_at ON uploads(status, updated_at);
CREATE INDEX IF NOT EXISTS idx_uploads_namespace_path ON uploads(namespace, path);

-- +goose Down
DROP INDEX IF EXISTS idx_uploads_namespace_path;
DROP INDEX IF EXISTS idx_uploads_status_updated_at;
DROP TABLE IF EXISTS uploads;
