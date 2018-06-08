package local

import (
	"reflect"
	"testing"
)

func TestProvider_Kind(t *testing.T) {
	t.Run("Storage Provider Kind", func(t *testing.T) {
		p := Provider{}
		if got := p.Kind(); got != "local" {
			t.Errorf("StorageProvider.Kind() = %v, Impossible, this should be local.", got)
		}
	})
}

func TestProvider_CollectionRef(t *testing.T) {
	tests := []struct {
		name string
		p    Provider
		want string
	}{
		{
			"Collection Reference",
			Provider{
				"./testdata",
				"subdir",
			},
			"subdir",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.CollectionRef(); got != tt.want {
				t.Errorf("Provider.CollectionRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_UploadFile(t *testing.T) {
	type args struct {
		filename  string
		reference string
	}
	tests := []struct {
		name    string
		p       Provider
		args    args
		wantErr bool
	}{
		{
			"Test Upload - upload.txt",
			Provider{
				"./testdata/dest_bucket",
				"subdir",
			},
			args{
				"./testdata/source_bucket/upload.txt",
				"upload.txt",
			},
			false,
		},
		{
			"Test Upload Bucket Error",
			Provider{
				"./testdata/dest_bucket",
				"subdir",
			},
			args{
				"does_not_exist.txt",
				"upload.txt",
			},
			true,
		},
		{
			"Test File Create Error",
			Provider{
				"./testdata/test_bucket",
				"subdir",
			},
			args{
				"./testdata/source_bucket/upload.txt",
				"",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.UploadFile(tt.args.filename, tt.args.reference); (err != nil) != tt.wantErr {
				t.Errorf("Provider.UploadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProvider_DownloadFile(t *testing.T) {
	type args struct {
		reference string
		filename  string
	}
	tests := []struct {
		name    string
		p       Provider
		args    args
		wantErr bool
	}{
		{
			"Test Upload - upload.txt",
			Provider{
				"./testdata/dest_bucket",
				"subdir",
			},
			args{
				"upload.txt",
				"./testdata/source_bucket/upload.txt",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.DownloadFile(tt.args.reference, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("Provider.DownloadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewLocalStorage(t *testing.T) {
	type args struct {
		serverPath string
		localPath string
	}
	tests := []struct {
		name string
		args args
		want *Provider
	}{
		{
			"Create Provider",
			args{
				"./testdata/dest_bucket",
				"subdir",
			},
			&Provider{
				"./testdata/dest_bucket",
				"subdir",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLocalStorage(tt.args.serverPath, tt.args.localPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLocalStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}