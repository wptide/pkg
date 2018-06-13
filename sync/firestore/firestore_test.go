package firestore

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/wptide/pkg/wporg"
	fsClient "github.com/wptide/pkg/wrapper/firestore"
)

type mockClient struct{}

func (m mockClient) GetDoc(path string) map[string]interface{} {

	switch path {
	case "test/get-sync-time":
		return map[string]interface{}{
			"mock-sync-start": int64(1526956796534182580),
		}
	case "test/set-sync-time":
		return map[string]interface{}{
			"mock-sync-start": int64(1526956796534182580),
		}
	case "test/set-sync-time-error":
		return nil
	case "test/get-sync-time-err":
		return nil
	case "test/update-check-error/plugin/test-project":
		return map[string]interface{}{
			"name": 12345,
		}
	default:
		return nil
	}
}

func (m mockClient) SetDoc(path string, data map[string]interface{}) error {
	return nil
}

func (m mockClient) AddDoc(collection string, data interface{}) error {
	return nil
}

func (m mockClient) DeleteDoc(path string) error {
	return nil
}

func (m mockClient) Close() error {
	return nil
}

func (m mockClient) Authenticated() bool {
	return true
}

func (m mockClient) QueryItems(collection string, conditions []fsClient.Condition, ordering []fsClient.Order, limit int, updateFunc fsClient.UpdateFunc) ([]interface{}, error) {
	return nil, nil
}

func TestFirestoreSync_UpdateCheck(t *testing.T) {
	type fields struct {
		ctx      context.Context
		client   fsClient.ClientInterface
		rootPath string
	}
	type args struct {
		project wporg.RepoProject
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"Update True",
			fields{
				context.Background(),
				&mockClient{},
				"test/update-check",
			},
			args{
				wporg.RepoProject{
					"Test Project",
					"test-project",
					"1.1.1",
					"2017-09-13 6:53pm GMT",
					"Short",
					"Long",
					"http://test.local",
					"plugin",
				},
			},
			true,
		},
		{
			"Update Check Error",
			fields{
				context.Background(),
				&mockClient{},
				"test/update-check-error",
			},
			args{
				wporg.RepoProject{
					"Test Project",
					"test-project",
					"1.1.1",
					"2017-09-13 6:53pm GMT",
					"Short",
					"Long",
					"http://test.local",
					"plugin",
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := NewWithClient(tt.fields.ctx, "mock-project", tt.fields.rootPath, tt.fields.client)
			if got := f.UpdateCheck(tt.args.project); got != tt.want {
				t.Errorf("FirestoreSync.UpdateCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFirestoreSync_RecordUpdate(t *testing.T) {
	type fields struct {
		ctx      context.Context
		client   fsClient.ClientInterface
		rootPath string
	}
	type args struct {
		project wporg.RepoProject
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Successful Update",
			fields{
				context.Background(),
				&mockClient{},
				"test/get-sync-time",
			},
			args{
				wporg.RepoProject{
					"Test Project",
					"test-project",
					"1.1.1",
					"2017-09-13 6:53pm GMT",
					"Short",
					"Long",
					"http://test.local",
					"plugin",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := NewWithClient(tt.fields.ctx, "mock-project", tt.fields.rootPath, tt.fields.client)
			if err := f.RecordUpdate(tt.args.project); (err != nil) != tt.wantErr {
				t.Errorf("FirestoreSync.RecordUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFirestoreSync_SetSyncTime(t *testing.T) {
	type fields struct {
		ctx      context.Context
		client   fsClient.ClientInterface
		rootPath string
	}
	type args struct {
		event       string
		projectType string
		t           time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"Set Time - No Error",
			fields{
				context.Background(),
				&mockClient{},
				"test/set-sync-time",
			},
			args{
				"start",
				"mock",
				time.Now(),
			},
		},
		{
			"Set Time - Error",
			fields{
				context.Background(),
				&mockClient{},
				"test/set-sync-time-error",
			},
			args{
				"start",
				"mock",
				time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := NewWithClient(tt.fields.ctx, "mock-project", tt.fields.rootPath, tt.fields.client)
			f.SetSyncTime(tt.args.event, tt.args.projectType, tt.args.t)
		})
	}
}

func TestFirestoreSync_GetSyncTime(t *testing.T) {

	type fields struct {
		ctx      context.Context
		client   fsClient.ClientInterface
		rootPath string
	}
	type args struct {
		event       string
		projectType string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   time.Time
	}{
		{
			"Get Time - No Error",
			fields{
				context.Background(),
				&mockClient{},
				"test/get-sync-time",
			},
			args{
				"start",
				"mock",
			},
			time.Unix(0, 1526956796534182580),
		},
		{
			"Get Time - Error",
			fields{
				context.Background(),
				&mockClient{},
				"test/get-sync-time-error",
			},
			args{
				"start",
				"mock",
			},
			time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := NewWithClient(tt.fields.ctx, "mock-project", tt.fields.rootPath, tt.fields.client)

			if got := f.GetSyncTime(tt.args.event, tt.args.projectType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FirestoreSync.GetSyncTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_itop(t *testing.T) {
	type args struct {
		data map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    wporg.RepoProject
		wantErr bool
	}{
		{
			"Map to Project",
			args{
				map[string]interface{}{
					"name":              "Test Project",
					"slug":              "test-project",
					"version":           "1.1.1",
					"last_updated":      "2017-09-13 6:53pm GMT",
					"short_description": "Short",
					"description":       "Long",
					"download_link":     "http://test.local",
					"type":              "plugin",
				},
			},
			wporg.RepoProject{
				"Test Project",
				"test-project",
				"1.1.1",
				"2017-09-13 6:53pm GMT",
				"Short",
				"Long",
				"http://test.local",
				"", // This is omitted and not required so will always be empty.
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := itop(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("itop() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("itop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ptoi(t *testing.T) {
	type args struct {
		project wporg.RepoProject
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			"Project to Map",
			args{
				wporg.RepoProject{
					"Test Project",
					"test-project",
					"1.1.1",
					"2017-09-13 6:53pm GMT",
					"Short",
					"Long",
					"http://test.local",
					"plugin",
				},
			},
			map[string]interface{}{
				"name":              "Test Project",
				"slug":              "test-project",
				"version":           "1.1.1",
				"last_updated":      "2017-09-13 6:53pm GMT",
				"short_description": "Short",
				"description":       "Long",
				"download_link":     "http://test.local",
				"type":              "plugin",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ptoi(tt.args.project)
			if (err != nil) != tt.wantErr {
				t.Errorf("ptoi() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ptoi() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		ctx         context.Context
		projectId   string
		rootDocPath string
	}
	tests := []struct {
		name string
		args args
		want reflect.Type
	}{
		{
			"Test New",
			args{
				context.Background(),
				"random-project",
				"sync/org.wordpress",
			},
			reflect.TypeOf(&FirestoreSync{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := New(tt.args.ctx, tt.args.projectId, tt.args.rootDocPath); reflect.TypeOf(got) != tt.want {
				t.Errorf("New() = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}

func TestNewWithClient(t *testing.T) {
	type args struct {
		ctx         context.Context
		projectId   string
		rootDocPath string
		client      fsClient.ClientInterface
	}
	tests := []struct {
		name string
		args args
		want reflect.Type
	}{
		{
			"New With Client",
			args{
				context.Background(),
				"random-id",
				"collection/doc",
				&mockClient{},
			},
			reflect.TypeOf(&FirestoreSync{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := NewWithClient(tt.args.ctx, tt.args.projectId, tt.args.rootDocPath, tt.args.client); reflect.TypeOf(got) != tt.want {
				t.Errorf("NewWithClient() = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}
