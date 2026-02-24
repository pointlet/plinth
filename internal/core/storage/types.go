package storage

import (
	"errors"
	"io"
	"time"
)

// ErrNotImplemented marks intentionally empty boilerplate methods.
var ErrNotImplemented = errors.New("storage: not implemented")

// Key identifies one logical object in storage.
type Key struct {
	Namespace string
	Path      string
}

// KeyPrefix is used for namespace-scoped listing.
type KeyPrefix struct {
	Namespace string
	Prefix    string
}

// UploadID identifies a resumable upload session.
type UploadID string

// UploadStatus tracks resumable upload lifecycle.
type UploadStatus string

const (
	UploadStatusActive    UploadStatus = "active"
	UploadStatusCommitted UploadStatus = "committed"
	UploadStatusAborted   UploadStatus = "aborted"
)

// StreamSource is the input stream for write operations.
type StreamSource interface {
	io.Reader
}

// ReadHandle is returned by Open and wraps stream + metadata.
type ReadHandle struct {
	Reader io.ReadCloser
	Info   ObjectInfo
}

// ObjectInfo is the canonical object metadata record.
type ObjectInfo struct {
	Key         Key
	Size        int64
	ETag        string
	ContentType string
	Filename    string
	CreatedAt   time.Time
	ModifiedAt  time.Time
	Version     string
	Deleted     bool
	Attributes  map[string]string
}

// PutOptions controls write semantics for Put.
type PutOptions struct {
	Filename       string
	ContentType    string
	Attributes     map[string]string
	ExpectAbsent   bool
	IfMatchETag    string
	IfMatchVersion string
	Sync           bool
}

// OpenOptions controls read behavior.
// Use Offset/Length for resumable reads (HTTP Range style).
type OpenOptions struct {
	Offset int64
	Length int64
}

// ListOptions controls paginated listing behavior.
type ListOptions struct {
	Limit  int
	Cursor string
}

// ListPage is a page of list results.
type ListPage struct {
	Items      []ObjectInfo
	NextCursor string
}

// DeleteOptions controls delete semantics.
type DeleteOptions struct {
	IfMatchETag    string
	IfMatchVersion string
}

// MoveOptions controls move semantics.
type MoveOptions struct {
	IfMatchETag    string
	IfMatchVersion string
}

// CopyOptions controls copy semantics.
type CopyOptions struct {
	IfMatchETag    string
	IfMatchVersion string
}

// CreateUploadOptions controls resumable upload session creation.
type CreateUploadOptions struct {
	Filename       string
	ContentType    string
	Attributes     map[string]string
	ExpectAbsent   bool
	IfMatchETag    string
	IfMatchVersion string
	ExpectedSize   int64
	Sync           bool
}

// AppendOptions controls one append operation to an upload session.
type AppendOptions struct {
	Offset int64
}

// AppendResult is returned after appending bytes to an upload session.
type AppendResult struct {
	Written       int64
	CurrentOffset int64
}

// CommitUploadOptions controls finalization semantics.
type CommitUploadOptions struct {
	Sync bool
}

// UploadInfo stores resumable upload session metadata.
type UploadInfo struct {
	ID           UploadID
	Key          Key
	Filename     string
	ContentType  string
	Attributes   map[string]string
	Status       UploadStatus
	TempPath     string
	BytesWritten int64
	ExpectedSize int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
