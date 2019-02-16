package firestore

import (
	"context"
	"reflect"
	"testing"

	"github.com/wptide/pkg/message"
	fsClient "github.com/wptide/pkg/wrapper/firestore"
)

func TestFirestoreProvider_SendMessage(t *testing.T) {
	ctx := context.Background()
	simpleClient, _ := NewWithClient(ctx, "client", "test", &mockClient{})
	failClient, _ := NewWithClient(ctx, "fail-client", "test-fail", &mockClient{})

	type args struct {
		msg *message.Message
	}
	tests := []struct {
		name    string
		fs      *Provider
		args    args
		wantErr bool
	}{
		{
			name: "Test Simple Message",
			fs:   simpleClient,
			args: args{
				&message.Message{
					Title: "Simple Message",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Failed Message",
			fs:   failClient,
			args: args{
				&message.Message{
					Title: "FAIL",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fs.SendMessage(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Provider.SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFirestoreProvider_GetNextMessage(t *testing.T) {
	ctx := context.Background()
	simpleClient, _ := NewWithClient(ctx, "mock-client", "simple-message", &mockClient{})
	lastRetryClient, _ := NewWithClient(ctx, "mock-client", "last-retry", &mockClient{})
	withIDClient, _ := NewWithClient(ctx, "mock-client", "with-id", &mockClient{})

	tests := []struct {
		name    string
		fs      *Provider
		want    *message.Message
		wantErr bool
	}{
		{
			name: "Get Simple Message",
			fs:   simpleClient,
			want: &message.Message{
				Title: "Simple Message",
			},
			wantErr: false,
		},
		{
			name: "Get Simple Message - Last Retry",
			fs:   lastRetryClient,
			want: &message.Message{
				Title: "Simple Message",
			},
			wantErr: false,
		},
		{
			name: "Get Simple Message - With ID",
			fs:   withIDClient,
			want: &message.Message{
				Title:       "Simple Message",
				ExternalRef: &[]string{"ABC123"}[0],
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fs.GetNextMessage()
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.GetNextMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.GetNextMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFirestoreProvider_DeleteMessage(t *testing.T) {
	ctx := context.Background()
	simpleClient, _ := NewWithClient(ctx, "mock-client", "delete-message", &mockClient{})

	type args struct {
		ref *string
	}
	tests := []struct {
		name    string
		fs      *Provider
		args    args
		wantErr bool
	}{
		{
			name: "Test Delete Message",
			fs:   simpleClient,
			args: args{
				&[]string{"BY7p9iOYjbT7Au4laiJ7"}[0],
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fs.DeleteMessage(tt.args.ref); (err != nil) != tt.wantErr {
				t.Errorf("Provider.DeleteMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		ctx         context.Context
		projectID   string
		rootDocPath string
	}
	tests := []struct {
		name string
		args args
		want reflect.Type
	}{
		{
			"Test New Client",
			args{
				context.Background(),
				"sample-project",
				"root-doc",
			},
			reflect.TypeOf(&Provider{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := New(tt.args.ctx, tt.args.projectID, tt.args.rootDocPath); reflect.TypeOf(got) != tt.want {
				t.Errorf("New() = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}

func TestNewWithClient(t *testing.T) {
	type args struct {
		ctx         context.Context
		projectID   string
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
				nil,
			},
			reflect.TypeOf(&Provider{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := NewWithClient(tt.args.ctx, tt.args.projectID, tt.args.rootDocPath, tt.args.client); reflect.TypeOf(got) != tt.want {
				t.Errorf("NewWithClient() = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}

//func TestNew(t *testing.T) {
//	type args struct {
//		ctx         context.Context
//		projectID   string
//		rootDocPath string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    reflect.Type
//		wantErr bool
//	}{
//		{
//			"Test New Client - Fail because API access",
//			args{
//				context.Background(),
//				"sample-project",
//				"root-doc",
//			},
//			reflect.TypeOf(&Provider{}),
//			true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got, _ := New(tt.args.ctx, tt.args.projectID, tt.args.rootDocPath); reflect.TypeOf(got) != tt.want {
//				t.Errorf("New() = %v, want %v", reflect.TypeOf(got), tt.want)
//			}
//		})
//	}
//}

func TestFirestoreProvider_Close(t *testing.T) {
	type fields struct {
		ctx      context.Context
		client   fsClient.ClientInterface
		rootPath string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"Close()",
			fields{
				context.Background(),
				&mockClient{},
				"",
			},
			false,
		},
		{
			"Close() - No client",
			fields{
				context.Background(),
				nil,
				"",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &Provider{
				ctx:      tt.fields.ctx,
				client:   tt.fields.client,
				rootPath: tt.fields.rootPath,
			}

			if err := fs.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Provider.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
