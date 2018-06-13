package gcs

import "io"

// StorageClient interface describes a new storage client.
type StorageClient interface {
	GetWriteCloser(bucket, ref string) (io.WriteCloser, error)
	GetReadCloser(bucket, ref string) (io.ReadCloser, error)
}
