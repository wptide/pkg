package gcs

import (
	"context"
	"io"
	"os"
)

var (
	fileCreate    = os.Create
	fileOpen      = os.Open
	storageObject = GSCClient(context.Background())
)

// Provider describes the GCS provider.
type Provider struct {
	ctx        context.Context
	client     *client
	projectID  *string
	bucketName *string
}

// Kind returns the kind of provider.
func (p Provider) Kind() string {
	return "gcs"
}


// CollectionRef returns an reference to a storage collection/bucket.
func (p Provider) CollectionRef() string {
	return *p.bucketName
}

// UploadFile puts the given file to the storage provider.
func (p Provider) UploadFile(filename, reference string) error {

	// Open file for writing to Cloud Storage.
	file, err := fileOpen(filename)

	// Error if file cannot be opened.
	if err != nil {
		return err
	}
	defer file.Close()

	w, _ := storageObject.GetWriteCloser(*p.bucketName, reference)
	defer w.Close()

	// Copy from file to object.
	if _, err := io.Copy(w, file); err != nil {
		return err
	}

	return nil
}

// DownloadFile gets the file from the storage provider.
func (p Provider) DownloadFile(reference, filename string) error {
	// Create file for writing.
	file, err := fileCreate(filename)

	// Error if file cannot be created.
	if err != nil {
		return err
	}
	defer file.Close()

	// Object to read from.
	r, err := storageObject.GetReadCloser(*p.bucketName, reference)
	defer r.Close()

	// Copy from object to file.
	if _, err := io.Copy(file, r); err != nil {
		return err
	}

	return nil
}

// NewCloudStorageProvider creates a new GCS provider.
func NewCloudStorageProvider(ctx context.Context, projectID string, bucketName string) *Provider {
	return &Provider{
		ctx:        ctx,
		projectID:  &projectID,
		bucketName: &bucketName,
	}
}
