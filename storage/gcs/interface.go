package gcs

import "io"

type StorageClient interface {
	GetWriteCloser(bucket, ref string) (io.WriteCloser, error)
	GetReadCloser(bucket, ref string) (io.ReadCloser, error)
}
