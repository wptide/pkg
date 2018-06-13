package gcs

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

// Provides a way to return an alternate objectHandle. Used for testing.
var objectWriterInterface = objectWriter
var objectReaderInterface = objectReader

// Interface which storage.Client implicitly implements.
type client interface {
	Bucket(name string) *storage.BucketHandle
}

type bucketHandle interface {
	Object(name string) *storage.ObjectHandle
}

type objectHandle interface {
	NewReader(ctx context.Context) (*storage.Reader, error)
	NewWriter(ctx context.Context) *storage.Writer
}

// Storage describes a new GCS client storage object.
type Storage struct {
	client client
	ctx    context.Context
}

func (s Storage) getBucket(bucketName string) bucketHandle {
	return s.client.Bucket(bucketName)
}

func (s Storage) getObject(bucket bucketHandle, ref string) objectHandle {
	return bucket.Object(ref)
}

func objectWriter(ctx context.Context, obj objectHandle) (io.WriteCloser, error) {
	w := obj.NewWriter(ctx)

	// Set object meta.
	w.ContentType = "application/json"
	w.Metadata = map[string]string{
		"x-goog-acl": "public-read",
	}

	return w, nil
}

func objectReader(ctx context.Context, obj objectHandle) (io.ReadCloser, error) {
	return obj.NewReader(ctx)
}

// GetWriteCloser gets a new io.WriteCloser for the storage client.
func (s *Storage) GetWriteCloser(bucket, ref string) (io.WriteCloser, error) {
	obj := s.getObject(s.getBucket(bucket), ref)
	return objectWriterInterface(s.ctx, obj)
}

// GetReadCloser gets a new io.ReadCloser for the storage client.
func (s *Storage) GetReadCloser(bucket, ref string) (io.ReadCloser, error) {
	obj := s.getObject(s.getBucket(bucket), ref)
	return objectReaderInterface(s.ctx, obj)
}

// GSCClient returns a new StorageClient.
func GSCClient(ctx context.Context) StorageClient {
	client, _ := storage.NewClient(ctx)

	return &Storage{
		client: client,
	}
}
