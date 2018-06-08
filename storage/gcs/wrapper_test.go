package gcs

import (
	"context"
	"io"
	"reflect"
	"testing"

	"cloud.google.com/go/storage"
)

type mockClient struct{}

func (m mockClient) Bucket(name string) *storage.BucketHandle {
	return &storage.BucketHandle{}
}

type mockBucket struct{}

func (m mockBucket) Object(name string) *storage.ObjectHandle {
	return &storage.ObjectHandle{}
}

type mockObject struct{}

func (m mockObject) NewWriter(ctx context.Context) *storage.Writer {
	return &storage.Writer{}
}

func (m mockObject) NewReader(ctx context.Context) (*storage.Reader, error) {
	return &storage.Reader{}, nil
}

type mockIO struct {
	readError  error
	writeError error
	s          string
	i          int64 // current reading index
	prevRune   int   // index of previous rune; or < 0
}

func (m mockIO) Write(p []byte) (n int, err error) {

	if m.writeError != nil {
		return 0, m.writeError
	}

	return len(p), nil
}

func (m mockIO) Close() error {
	return nil
}

func (m mockIO) Read(b []byte) (n int, err error) {

	if m.readError != nil {
		return 0, m.readError
	}

	if m.i >= int64(len(m.s)) {
		return 0, io.EOF
	}
	m.prevRune = -1
	n = copy(b, m.s[m.i:])
	m.i += int64(n)
	return
}

func mockWriterInterface(ctx context.Context, obj objectHandle) (io.WriteCloser, error) {
	return &mockIO{}, nil
}

func mockReaderInterface(ctx context.Context, obj objectHandle) (io.ReadCloser, error) {
	return &mockIO{}, nil
}

func TestStorage_getBucket(t *testing.T) {
	type fields struct {
		client client
		ctx    context.Context
	}
	type args struct {
		bucketName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bucketHandle
	}{
		{
			"Get bucket",
			fields{
				&mockClient{},
				context.Background(),
			},
			args{
				bucketName: "arbitrary",
			},
			&storage.BucketHandle{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Storage{
				client: tt.fields.client,
				ctx:    tt.fields.ctx,
			}
			if got := s.getBucket(tt.args.bucketName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.getBucket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_getObject(t *testing.T) {
	type fields struct {
		client client
		ctx    context.Context
	}
	type args struct {
		bucket bucketHandle
		ref    string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   objectHandle
	}{
		{
			"Get object",
			fields{
				&mockClient{},
				context.Background(),
			},
			args{
				&mockBucket{},
				"arbitrary.name",
			},
			&storage.ObjectHandle{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Storage{
				client: tt.fields.client,
				ctx:    tt.fields.ctx,
			}
			if got := s.getObject(tt.args.bucket, tt.args.ref); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.getObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_GetWriteCloser(t *testing.T) {

	oldWriterFunc := objectWriterInterface
	objectWriterInterface = mockWriterInterface
	defer func() {
		objectWriterInterface = oldWriterFunc
	}()

	type fields struct {
		client client
		ctx    context.Context
	}
	type args struct {
		bucket string
		ref    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    io.WriteCloser
		wantErr bool
	}{
		{
			"Get Write Closer",
			fields{
				&mockClient{},
				context.Background(),
			},
			args{
				"test_bucket",
				"test_ref",
			},
			&mockIO{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				client: tt.fields.client,
				ctx:    tt.fields.ctx,
			}
			got, err := s.GetWriteCloser(tt.args.bucket, tt.args.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.GetWriteCloser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.GetWriteCloser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_GetReadCloser(t *testing.T) {

	oldReaderFunc := objectReaderInterface
	objectReaderInterface = mockReaderInterface
	defer func() {
		objectReaderInterface = oldReaderFunc
	}()

	type fields struct {
		client client
		ctx    context.Context
	}
	type args struct {
		bucket string
		ref    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    io.ReadCloser
		wantErr bool
	}{
		{
			"Get Read Closer",
			fields{
				&mockClient{},
				context.Background(),
			},
			args{
				"test_bucket",
				"test_ref",
			},
			&mockIO{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				client: tt.fields.client,
				ctx:    tt.fields.ctx,
			}
			got, err := s.GetReadCloser(tt.args.bucket, tt.args.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.GetReadCloser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.GetReadCloser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_objectWriter(t *testing.T) {
	type args struct {
		ctx context.Context
		obj objectHandle
	}
	tests := []struct {
		name    string
		args    args
		want    reflect.Type
		wantErr bool
	}{
		{
			"objectWriter",
			args{
				context.Background(),
				&mockObject{},
			},
			reflect.TypeOf(&storage.Writer{}),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := objectWriter(tt.args.ctx, tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("objectWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(reflect.TypeOf(got), tt.want) {
				t.Errorf("objectWriter() = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}

func Test_objectReader(t *testing.T) {
	type args struct {
		ctx context.Context
		obj objectHandle
	}
	tests := []struct {
		name    string
		args    args
		want    reflect.Type
		wantErr bool
	}{
		{
			"objectReader",
			args{
				context.Background(),
				&mockObject{},
			},
			reflect.TypeOf(&storage.Reader{}),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := objectReader(tt.args.ctx, tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("objectReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(reflect.TypeOf(got), tt.want) {
				t.Errorf("objectReader() = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}
