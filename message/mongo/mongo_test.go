package mongo

import (
	"context"
	"fmt"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/wptide/pkg/message"
	wrapper "github.com/wptide/pkg/wrapper/mongo"
)

func testServer(t *testing.T, handler func(net.Conn)) (func() error, string) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")

	go func() {
		for {
			// Listen for an incoming connection.
			server, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				os.Exit(1)
			}
			// Handle connections in a new goroutine.
			if handler != nil {
				go handler(server)
			}
		}
	}()

	return ln.Close, ln.Addr().String()
}

func TestMongoProvider_SendMessage(t *testing.T) {
	type fields struct {
		ctx        context.Context
		client     wrapper.Client
		database   string
		collection string
	}
	type args struct {
		msg *message.Message
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Send Message",
			fields{
				context.Background(),
				&MockClient{},
				"test-db",
				"test-collection",
			},
			args{
				&message.Message{
					Title: "Test Plugin",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewWithClient(tt.fields.ctx, tt.fields.database, tt.fields.collection, tt.fields.client)
			if err := m.SendMessage(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("MongoProvider.SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMongoProvider_GetNextMessage(t *testing.T) {
	type fields struct {
		ctx        context.Context
		client     wrapper.Client
		database   string
		collection string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *message.Message
		wantErr bool
	}{
		{
			"Get Next Message - No Records",
			fields{
				context.Background(),
				&MockClient{
					"test-no-records",
				},
				"test",
				"test-no-records",
			},
			nil,
			true,
		},
		{
			"Get Next Message - Valid Message",
			fields{
				context.Background(),
				&MockClient{
					"test-valid-message",
				},
				"test",
				"test-valid-message",
			},
			&message.Message{
				Title:       "Plugin One",
				ExternalRef: &[]string{"abcdef123456789009876364"}[0],
			},
			false,
		},
		{
			"Get Next Message - Valid Message - No Retry",
			fields{
				context.Background(),
				&MockClient{
					"test-valid-message-no-retry",
				},
				"test",
				"test-valid-message-no-retry",
			},
			&message.Message{
				Title:       "Plugin One",
				ExternalRef: &[]string{"abcdef123456789009876364"}[0],
			},
			false,
		},
		{
			"Get Next Message - Lock Fail",
			fields{
				context.Background(),
				&MockClient{
					"test-lock-fail",
				},
				"test",
				"test-lock-fail",
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewWithClient(tt.fields.ctx, tt.fields.database, tt.fields.collection, tt.fields.client)

			got, err := m.GetNextMessage()
			if (err != nil) != tt.wantErr {
				t.Errorf("MongoProvider.GetNextMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MongoProvider.GetNextMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoProvider_DeleteMessage(t *testing.T) {
	type fields struct {
		ctx        context.Context
		client     wrapper.Client
		database   string
		collection string
	}
	type args struct {
		ref *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Delete Message",
			fields{
				context.Background(),
				&MockClient{},
				"test-db",
				"test-collection",
			},
			args{
				&[]string{"mock-ref"}[0],
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewWithClient(tt.fields.ctx, tt.fields.database, tt.fields.collection, tt.fields.client)
			if err := m.DeleteMessage(tt.args.ref); (err != nil) != tt.wantErr {
				t.Errorf("MongoProvider.DeleteMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	_, host := testServer(t, nil)

	type args struct {
		ctx        context.Context
		user       string
		pass       string
		host       string
		db         string
		collection string
		opts       *mongo.ClientOptions
	}
	tests := []struct {
		name    string
		args    args
		want    reflect.Type
		wantErr bool
	}{
		{
			"New Mongo Client",
			args{
				context.Background(),
				"",
				"",
				host,
				"database",
				"collection",
				nil,
			},
			reflect.TypeOf(&MongoProvider{}),
			false,
		},
		{
			"New Mongo Client - Host Err",
			args{
				context.Background(),
				"",
				"",
				"",
				"database",
				"collection",
				nil,
			},
			reflect.TypeOf(&MongoProvider{}),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.ctx, tt.args.user, tt.args.pass, tt.args.host, tt.args.db, tt.args.collection, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != tt.want {
				t.Errorf("New() = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}

func TestMongoProvider_Close(t *testing.T) {
	type fields struct {
		ctx        context.Context
		client     wrapper.Client
		database   string
		collection string
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
				&MockClient{},
				"test-db",
				"test-col",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewWithClient(tt.fields.ctx, tt.fields.database, tt.fields.collection, tt.fields.client)
			if err := m.Close(); (err != nil) != tt.wantErr {
				t.Errorf("MongoProvider.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
