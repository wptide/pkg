package gcs

import (
	"cloud.google.com/go/storage"
	"context"
	"io"
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

func (s *Storage) GetWriteCloser(bucket, ref string) (io.WriteCloser, error) {
	obj := s.getObject(s.getBucket(bucket), ref)
	return objectWriterInterface(s.ctx, obj)
}

func (s *Storage) GetReadCloser(bucket, ref string) (io.ReadCloser, error) {
	obj := s.getObject(s.getBucket(bucket), ref)
	return objectReaderInterface(s.ctx, obj)
}

func GSCClient(ctx context.Context) StorageClient {
	client, _ := storage.NewClient(ctx)

	return &Storage{
		client: client,
	}
}
