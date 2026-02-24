package storage

import "context"

// ObjectStore owns committed object-record persistence, independent from blob bytes.
// Implementations are backend-agnostic (SQLite, etc.).
type ObjectStore interface {
	Upsert(requestContext context.Context, objectInfo ObjectInfo) error
	Get(requestContext context.Context, key Key) (ObjectInfo, error)
	List(requestContext context.Context, prefix KeyPrefix, options ListOptions) (ListPage, error)
	Delete(requestContext context.Context, key Key, options DeleteOptions) error
	Move(requestContext context.Context, sourceKey Key, destinationKey Key, options MoveOptions) (ObjectInfo, error)
	Copy(requestContext context.Context, sourceKey Key, destinationKey Key, options CopyOptions) (ObjectInfo, error)
}

// UploadStore owns resumable upload session metadata.
// This is backend-agnostic and can be implemented by SQLite or any other DB.
type UploadStore interface {
	CreateUpload(requestContext context.Context, uploadInfo UploadInfo) (UploadInfo, error)
	GetUpload(requestContext context.Context, uploadID UploadID) (UploadInfo, error)
	UpdateUploadProgress(requestContext context.Context, uploadID UploadID, currentOffset int64) (UploadInfo, error)
	CommitUpload(requestContext context.Context, uploadID UploadID, finalObjectInfo ObjectInfo) error
	AbortUpload(requestContext context.Context, uploadID UploadID) error
}
