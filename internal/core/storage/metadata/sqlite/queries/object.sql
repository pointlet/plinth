-- name: UpsertObject :exec
INSERT INTO objects (
  namespace,
  path,
  size,
  etag,
  content_type,
  filename,
  version,
  deleted,
  attributes_json,
  created_at,
  modified_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(namespace, path) DO UPDATE SET
  size = excluded.size,
  etag = excluded.etag,
  content_type = excluded.content_type,
  filename = excluded.filename,
  version = excluded.version,
  deleted = excluded.deleted,
  attributes_json = excluded.attributes_json,
  modified_at = excluded.modified_at;

-- name: GetObject :one
SELECT
  namespace,
  path,
  size,
  etag,
  content_type,
  filename,
  version,
  deleted,
  attributes_json,
  created_at,
  modified_at
FROM objects
WHERE objects.namespace = ? AND objects.path = ?;

-- name: ListObjectsByPrefix :many
SELECT
  namespace,
  path,
  size,
  etag,
  content_type,
  filename,
  version,
  deleted,
  attributes_json,
  created_at,
  modified_at
FROM objects
WHERE namespace = ?
  AND path LIKE ? || '%'
ORDER BY path
LIMIT ? OFFSET ?;

-- name: DeleteObject :exec
DELETE FROM objects
WHERE namespace = ? AND path = ?;

-- name: MoveObjectRecord :exec
UPDATE objects
SET namespace = ?, path = ?, modified_at = ?
WHERE namespace = ? AND path = ?;

-- name: CopyObjectRecord :exec
INSERT INTO objects (
  namespace,
  path,
  size,
  etag,
  content_type,
  filename,
  version,
  deleted,
  attributes_json,
  created_at,
  modified_at
)
SELECT
  ?, ?, size, etag, content_type, filename, ?, deleted, attributes_json, ?, ?
FROM objects
WHERE objects.namespace = ? AND objects.path = ?;
