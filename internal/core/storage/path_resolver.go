package storage

import "context"

// PathResolver maps logical keys to safe, rooted filesystem paths.
type PathResolver interface {
	ResolveObjectPath(requestContext context.Context, key Key) (string, error)
	ResolvePrefixPath(requestContext context.Context, keyPrefix KeyPrefix) (string, error)
	ResolveUploadTempPath(requestContext context.Context, uploadID UploadID) (string, error)
}

// FSPathResolver is a local-filesystem resolver skeleton.
type FSPathResolver struct {
	root string
}

func NewFSPathResolver(root string) *FSPathResolver {
	return &FSPathResolver{root: root}
}

func (r *FSPathResolver) ResolveObjectPath(requestContext context.Context, key Key) (string, error) {
	return "", ErrNotImplemented
}

func (r *FSPathResolver) ResolvePrefixPath(requestContext context.Context, keyPrefix KeyPrefix) (string, error) {
	return "", ErrNotImplemented
}

func (r *FSPathResolver) ResolveUploadTempPath(requestContext context.Context, uploadID UploadID) (string, error) {
	return "", ErrNotImplemented
}
