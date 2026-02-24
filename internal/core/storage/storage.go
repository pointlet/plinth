// Package storage defines Plinth's core storage contracts and local engine wiring.
package storage

import "context"

// Engine is the storage contract exposed to core components.
// TODO: make sure all operations are stream-friendly and key-based.
type Engine interface {
	Put(ctx context.Context, key Key, src StreamSource, opt PutOptions) (ObjectInfo, error)
	Open(ctx context.Context, key Key, opt OpenOptions) (ReadHandle, error)
	Stat(ctx context.Context, key Key) (ObjectInfo, error)
	List(ctx context.Context, prefix KeyPrefix, opt ListOptions) (ListPage, error)
	Delete(ctx context.Context, key Key, opt DeleteOptions) error
	Move(ctx context.Context, src Key, dst Key, opt MoveOptions) (ObjectInfo, error)
	Copy(ctx context.Context, src Key, dst Key, opt CopyOptions) (ObjectInfo, error)

	CreateUpload(ctx context.Context, key Key, opt CreateUploadOptions) (UploadInfo, error)
	AppendUpload(ctx context.Context, uploadID UploadID, src StreamSource, opt AppendOptions) (AppendResult, error)
	StatUpload(ctx context.Context, uploadID UploadID) (UploadInfo, error)
	CommitUpload(ctx context.Context, uploadID UploadID, opt CommitUploadOptions) (ObjectInfo, error)
	AbortUpload(ctx context.Context, uploadID UploadID) error
}

// LocalEngine is the filesystem-backed engine coordinator.
// It composes resolver, blob writer, object store, and upload store helpers.
type LocalEngine struct {
	resolver PathResolver
	writer   BlobWriter
	objects  ObjectStore
	uploads  UploadStore
}

// New creates a local engine skeleton
func New(resolver PathResolver, writer BlobWriter, objects ObjectStore, uploads UploadStore) *LocalEngine {
	return &LocalEngine{
		resolver: resolver,
		writer:   writer,
		objects:  objects,
		uploads:  uploads,
	}
}

func (e *LocalEngine) Put(ctx context.Context, key Key, src StreamSource, opt PutOptions) (ObjectInfo, error) {
	return ObjectInfo{}, ErrNotImplemented
}

func (e *LocalEngine) Open(ctx context.Context, key Key, opt OpenOptions) (ReadHandle, error) {
	return ReadHandle{}, ErrNotImplemented
}

func (e *LocalEngine) Stat(ctx context.Context, key Key) (ObjectInfo, error) {
	return ObjectInfo{}, ErrNotImplemented
}

func (e *LocalEngine) List(ctx context.Context, prefix KeyPrefix, opt ListOptions) (ListPage, error) {
	return ListPage{}, ErrNotImplemented
}

func (e *LocalEngine) Delete(ctx context.Context, key Key, opt DeleteOptions) error {
	return ErrNotImplemented
}

func (e *LocalEngine) Move(ctx context.Context, src Key, dst Key, opt MoveOptions) (ObjectInfo, error) {
	return ObjectInfo{}, ErrNotImplemented
}

func (e *LocalEngine) Copy(ctx context.Context, src Key, dst Key, opt CopyOptions) (ObjectInfo, error) {
	return ObjectInfo{}, ErrNotImplemented
}

func (e *LocalEngine) CreateUpload(ctx context.Context, key Key, opt CreateUploadOptions) (UploadInfo, error) {
	return UploadInfo{}, ErrNotImplemented
}

func (e *LocalEngine) AppendUpload(ctx context.Context, uploadID UploadID, src StreamSource, opt AppendOptions) (AppendResult, error) {
	return AppendResult{}, ErrNotImplemented
}

func (e *LocalEngine) StatUpload(ctx context.Context, uploadID UploadID) (UploadInfo, error) {
	return UploadInfo{}, ErrNotImplemented
}

func (e *LocalEngine) CommitUpload(ctx context.Context, uploadID UploadID, opt CommitUploadOptions) (ObjectInfo, error) {
	return ObjectInfo{}, ErrNotImplemented
}

func (e *LocalEngine) AbortUpload(ctx context.Context, uploadID UploadID) error {
	return ErrNotImplemented
}
