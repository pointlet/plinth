package storage

import (
	"context"
	"io"
)

// BlobWriter combines single-shot and resumable write primitives.
type BlobWriter interface {
	AtomicWriter
	ResumableWriter
}

// AtomicWriter performs durable stream writes through a temp-file commit path.
type AtomicWriter interface {
	WriteAtomic(requestContext context.Context, destinationAbsolutePath string, sourceStream io.Reader, options AtomicWriteOptions) (AtomicWriteResult, error)
}

// ResumableWriter manages append/commit/abort operations for upload temp files.
type ResumableWriter interface {
	AppendAt(requestContext context.Context, temporaryAbsolutePath string, expectedOffset int64, sourceStream io.Reader, options AppendWriteOptions) (AppendWriteResult, error)
	CommitTempFile(requestContext context.Context, temporaryAbsolutePath string, finalAbsolutePath string, options CommitWriteOptions) (AtomicWriteResult, error)
	AbortTempFile(requestContext context.Context, temporaryAbsolutePath string) error
}

// AtomicWriteOptions controls durability and precondition checks.
type AtomicWriteOptions struct {
	Sync         bool
	ExpectAbsent bool
}

// AtomicWriteResult contains write-path information needed by metadata logic.
type AtomicWriteResult struct {
	Size          int64
	ETag          string
	CreatedAtUnix int64
	UpdatedAtUnix int64
}

// AppendWriteOptions controls low-level append behavior.
type AppendWriteOptions struct {
	Sync bool
}

// AppendWriteResult is returned by low-level append operations.
type AppendWriteResult struct {
	Written       int64
	CurrentOffset int64
}

// CommitWriteOptions controls low-level commit behavior.
type CommitWriteOptions struct {
	Sync bool
}

// FileAtomicWriter is a filesystem-backed atomic writer skeleton.
type FileAtomicWriter struct{}

func NewFileAtomicWriter() *FileAtomicWriter {
	return &FileAtomicWriter{}
}

func (w *FileAtomicWriter) WriteAtomic(requestContext context.Context, destinationAbsolutePath string, sourceStream io.Reader, options AtomicWriteOptions) (AtomicWriteResult, error) {
	return AtomicWriteResult{}, ErrNotImplemented
}

func (w *FileAtomicWriter) AppendAt(requestContext context.Context, temporaryAbsolutePath string, expectedOffset int64, sourceStream io.Reader, options AppendWriteOptions) (AppendWriteResult, error) {
	return AppendWriteResult{}, ErrNotImplemented
}

func (w *FileAtomicWriter) CommitTempFile(requestContext context.Context, temporaryAbsolutePath string, finalAbsolutePath string, options CommitWriteOptions) (AtomicWriteResult, error) {
	return AtomicWriteResult{}, ErrNotImplemented
}

func (w *FileAtomicWriter) AbortTempFile(requestContext context.Context, temporaryAbsolutePath string) error {
	return ErrNotImplemented
}
