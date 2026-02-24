package sqlite

import (
	"context"

	"github.com/pointlet/plinth/internal/core/storage"
)

func (store *Store) CreateUpload(requestContext context.Context, uploadInfo storage.UploadInfo) (storage.UploadInfo, error) {
	return storage.UploadInfo{}, storage.ErrNotImplemented
}

func (store *Store) GetUpload(requestContext context.Context, uploadID storage.UploadID) (storage.UploadInfo, error) {
	return storage.UploadInfo{}, storage.ErrNotImplemented
}

func (store *Store) UpdateUploadProgress(requestContext context.Context, uploadID storage.UploadID, currentOffset int64) (storage.UploadInfo, error) {
	return storage.UploadInfo{}, storage.ErrNotImplemented
}

func (store *Store) CommitUpload(requestContext context.Context, uploadID storage.UploadID, finalObjectInfo storage.ObjectInfo) error {
	return storage.ErrNotImplemented
}

func (store *Store) AbortUpload(requestContext context.Context, uploadID storage.UploadID) error {
	return storage.ErrNotImplemented
}
