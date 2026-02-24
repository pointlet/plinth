package sqlite

import (
	"context"

	"github.com/pointlet/plinth/internal/core/storage"
	"github.com/pointlet/plinth/internal/core/storage/metadata/sqlite/sqlcgen"
)

func (store *Store) Upsert(requestContext context.Context, objectInfo storage.ObjectInfo) error {
	return storage.ErrNotImplemented
}

func (store *Store) Get(requestContext context.Context, key storage.Key) (storage.ObjectInfo, error) {
	row, err := store.generatedQueries.GetObject(requestContext, sqlcgen.GetObjectParams{
		Namespace: key.Namespace,
		Path:      key.Path,
	})
	if err != nil {
		return storage.ObjectInfo{}, err
	}

	return storage.ObjectInfo{
		Key: storage.Key{
			Namespace: row.Namespace,
			Path:      row.Path,
		},
		Size:        row.Size,
		ETag:        row.Etag,
		ContentType: row.ContentType,
		Filename:    row.Filename,
		Version:     row.Version,
		Deleted:     row.Deleted != 0,
	}, nil
}

func (store *Store) List(requestContext context.Context, prefix storage.KeyPrefix, options storage.ListOptions) (storage.ListPage, error) {
	return storage.ListPage{}, storage.ErrNotImplemented
}

func (store *Store) Delete(requestContext context.Context, key storage.Key, options storage.DeleteOptions) error {
	return storage.ErrNotImplemented
}

func (store *Store) Move(requestContext context.Context, sourceKey storage.Key, destinationKey storage.Key, options storage.MoveOptions) (storage.ObjectInfo, error) {
	return storage.ObjectInfo{}, storage.ErrNotImplemented
}

func (store *Store) Copy(requestContext context.Context, sourceKey storage.Key, destinationKey storage.Key, options storage.CopyOptions) (storage.ObjectInfo, error) {
	return storage.ObjectInfo{}, storage.ErrNotImplemented
}
