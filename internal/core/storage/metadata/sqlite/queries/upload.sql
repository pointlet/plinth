-- name: CreateUploadRecord :exec
INSERT INTO uploads (
  upload_id,
  namespace,
  path,
  filename,
  content_type,
  attributes_json,
  temp_path,
  bytes_written,
  expected_size,
  status,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetUploadRecord :one
SELECT
  upload_id,
  namespace,
  path,
  filename,
  content_type,
  attributes_json,
  temp_path,
  bytes_written,
  expected_size,
  status,
  created_at,
  updated_at
FROM uploads
WHERE upload_id = ?;

-- name: UpdateUploadProgress :exec
UPDATE uploads
SET bytes_written = ?, updated_at = ?
WHERE upload_id = ?
  AND status = 'active';

-- name: MarkUploadCommitted :exec
UPDATE uploads
SET status = 'committed', updated_at = ?
WHERE upload_id = ?
  AND status = 'active';

-- name: MarkUploadAborted :exec
UPDATE uploads
SET status = 'aborted', updated_at = ?
WHERE upload_id = ?
  AND status = 'active';
