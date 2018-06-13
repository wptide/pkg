package gcs

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	"cloud.google.com/go/storage"
)

type mockStorageClient struct{}

func (m mockStorageClient) GetWriteCloser(bucket, ref string) (io.WriteCloser, error) {

	switch ref {
	case "bucket_error.txt":
		return &mockIO{
			writeError: errors.New("bucket error"),
		}, errors.New("bucket error")
	default:
		return &mockIO{}, nil
	}
}

func (m mockStorageClient) GetReadCloser(bucket, ref string) (io.ReadCloser, error) {

	switch ref {
	case "bucket_error.txt":
		return &mockIO{
			readError: errors.New("bucket error"),
		}, errors.New("bucket error")
	default:
		return &mockIO{}, nil
	}
}

func mockFileOpen(name string) (*os.File, error) {
	switch name {
	case "error.txt":
		return nil, errors.New("something went wrong")
	default:
		return os.Open("./testdata/raw.txt")
	}
}

func mockFileCreate(name string) (*os.File, error) {
	switch name {
	case "error.txt":
		return nil, errors.New("something went wrong")
	default:
		//return os.Create("./testdata/output.txt")
		return nil, nil
		//strings.Reader{}
	}
}

func TestProvider_Kind(t *testing.T) {
	t.Run("Storage Provider Kind", func(t *testing.T) {
		p := Provider{}
		if got := p.Kind(); got != "gcs" {
			t.Errorf("Provider.Kind() = %v, Impossible, this should be gcloud/storage.", got)
		}
	})
}

func TestProvider_CollectionRef(t *testing.T) {
	type fields struct {
		ctx        context.Context
		client     *storage.Client
		projectID  *string
		bucketName *string
		bucket     *storage.BucketHandle
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"Collection Reference",
			fields{
				bucketName: &[]string{"test_bucket"}[0],
			},
			"test_bucket",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Provider{
				bucketName: tt.fields.bucketName,
			}
			if got := p.CollectionRef(); got != tt.want {
				t.Errorf("Provider.CollectionRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_UploadFile(t *testing.T) {

	// Set storage object.
	storageObject = &mockStorageClient{}
	defer func() { storageObject = GSCClient(context.Background()) }()

	// Set out fileOpen variable to the mock function.
	fileOpen = mockFileOpen
	// Remember to set it back after the test.
	defer func() { fileOpen = os.Open }()

	type fields struct {
		ctx           context.Context
		client        *client
		projectID     *string
		bucketName    *string
		storageObject StorageClient
	}
	type args struct {
		filename  string
		reference string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Test Upload - upload.txt",
			fields{
				ctx:        context.Background(),
				bucketName: &[]string{"testBucket"}[0],
			},
			args{
				"upload.txt",
				"upload.txt",
			},
			false,
		},
		{
			"Test Upload Bucket Error",
			fields{
				ctx:        context.Background(),
				bucketName: &[]string{"testBucket"}[0],
			},
			args{
				"bucket_error.txt",
				"bucket_error.txt",
			},
			true,
		},
		{
			"Test File Open Error",
			fields{
				ctx:        context.Background(),
				bucketName: &[]string{"testBucket"}[0],
			},
			args{
				"error.txt",
				"error.txt",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Provider{
				ctx:        tt.fields.ctx,
				client:     tt.fields.client,
				projectID:  tt.fields.projectID,
				bucketName: tt.fields.bucketName,
			}
			if err := p.UploadFile(tt.args.filename, tt.args.reference); (err != nil) != tt.wantErr {
				t.Errorf("Provider.UploadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCloudStorageProvider(t *testing.T) {

	ctx := context.Background()

	//oldClient := storageClient
	//storageClient = &mockStorageClient{}
	//defer func() {
	//	storageClient = oldClient
	//}()

	type args struct {
		ctx        context.Context
		projectId  string
		bucketName string
	}
	tests := []struct {
		name string
		args args
		want reflect.Type
	}{
		{"Test GCloud Storage Provider",
			args{
				ctx,
				"stide-development-201405",
				"stide-development-201405.appspot.com",
			},
			reflect.TypeOf(&Provider{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCloudStorageProvider(tt.args.ctx, tt.args.projectId, tt.args.bucketName); !reflect.DeepEqual(reflect.TypeOf(got), tt.want) {
				t.Errorf("NewCloudStorageProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_DownloadFile(t *testing.T) {

	// Set storage object.
	storageObject = &mockStorageClient{}
	defer func() { storageObject = GSCClient(context.Background()) }()

	// Set out fileOpen variable to the mock function.
	fileCreate = mockFileCreate
	// Remember to set it back after the test.
	defer func() { fileCreate = os.Create }()

	type fields struct {
		ctx        context.Context
		client     *client
		projectID  *string
		bucketName *string
	}
	type args struct {
		reference string
		filename  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Test Download - download.txt",
			fields{
				ctx:        context.Background(),
				bucketName: &[]string{"testBucket"}[0],
			},
			args{
				"download.txt",
				"download.txt",
			},
			false,
		},
		{
			"Test File Creare Error",
			fields{
				ctx:        context.Background(),
				bucketName: &[]string{"testBucket"}[0],
			},
			args{
				"error.txt",
				"error.txt",
			},
			true,
		},
		{
			"Test Bucket File Error",
			fields{
				ctx:        context.Background(),
				bucketName: &[]string{"testBucket"}[0],
			},
			args{
				"bucket_error.txt",
				"bucket_error.txt",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Provider{
				ctx:        tt.fields.ctx,
				client:     tt.fields.client,
				projectID:  tt.fields.projectID,
				bucketName: tt.fields.bucketName,
			}
			if err := p.DownloadFile(tt.args.reference, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("Provider.DownloadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
