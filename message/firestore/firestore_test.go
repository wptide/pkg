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
		fs      *FirestoreProvider
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
				t.Errorf("FirestoreProvider.SendMessage() error = %v, wantErr %v", err, tt.wantErr)
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
		fs      *FirestoreProvider
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
				t.Errorf("FirestoreProvider.GetNextMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FirestoreProvider.GetNextMessage() = %v, want %v", got, tt.want)
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
		fs      *FirestoreProvider
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
				t.Errorf("FirestoreProvider.DeleteMessage() error = %v, wantErr %v", err, tt.wantErr)
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
		name    string
		args    args
		want    *FirestoreProvider
		wantErr bool
	}{
		{
			"Test New Client - Fail because API acces",
			args{
				context.Background(),
				"sample-project",
				"root-doc",
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.ctx, tt.args.projectId, tt.args.rootDocPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			fs := &FirestoreProvider{
				ctx:      tt.fields.ctx,
				client:   tt.fields.client,
				rootPath: tt.fields.rootPath,
			}

			if err := fs.Close(); (err != nil) != tt.wantErr {
				t.Errorf("FirestoreProvider.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
