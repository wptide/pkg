package mongo

import (
	"context"
	"fmt"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/mongodb/mongo-go-driver/core/option"
	"github.com/mongodb/mongo-go-driver/mongo"
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

func TestNewMongoClient(t *testing.T) {
	type args struct {
		user string
		pass string
		host string
		opts *mongo.ClientOptions
	}
	tests := []struct {
		name    string
		args    args
		want    reflect.Type
		wantErr bool
	}{
		{
			"Mongo Client",
			args{
				"",
				"",
				"localhost:27017",
				nil,
			},
			reflect.TypeOf(&MongoClient{}),
			false,
		},
		{
			"Mongo Client - With User",
			args{
				"root",
				"",
				"localhost:27017",
				nil,
			},
			reflect.TypeOf(&MongoClient{}),
			false,
		},
		{
			"Mongo Client - With User and Pass",
			args{
				"root",
				"password",
				"localhost:27017",
				nil,
			},
			reflect.TypeOf(&MongoClient{}),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMongoClient(context.Background(), tt.args.user, tt.args.pass, tt.args.host, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMongoClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != tt.want {
				t.Errorf("NewMongoClient() = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}

func TestMongoClient_Database(t *testing.T) {

	_, host := testServer(t, nil)
	client, _ := mongo.NewClient("mongodb://" + host)

	type fields struct {
		Client *mongo.Client
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   reflect.Type
	}{
		{
			"Database()",
			fields{
				client,
			},
			args{
				"test_db",
			},
			reflect.TypeOf(&MongoDatabase{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := MongoClient{
				Client: tt.fields.Client,
			}
			if got := mc.Database(tt.args.name); reflect.TypeOf(got) != tt.want {
				t.Errorf("MongoClient.Database() = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}

func TestMongoDatabase_Collection(t *testing.T) {
	type fields struct {
		Database *mongo.Database
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   reflect.Type
	}{
		{
			"Collection()",
			fields{
				&mongo.Database{},
			},
			args{
				"test_collection",
			},
			reflect.TypeOf(&MongoCollection{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := MongoDatabase{
				Database: tt.fields.Database,
			}
			if got := d.Collection(tt.args.name); reflect.TypeOf(got) != tt.want {
				t.Errorf("MongoDatabase.Collection() = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}

func TestMongoCollection_InsertOne(t *testing.T) {
	type fields struct {
		Collection CollectionLayer
	}
	type args struct {
		ctx      context.Context
		document interface{}
		opts     []option.InsertOneOptioner
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    InsertOneResultLayer
		wantErr bool
	}{
		{
			"InsertOne()",
			fields{
				&MongoCollection{},
			},
			args{
				context.Background(),
				map[string]interface{}{},
				nil,
			},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields.Collection
			got, err := c.InsertOne(tt.args.ctx, tt.args.document, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("MongoCollection.InsertOne() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MongoCollection.InsertOne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoCollection_FindOne(t *testing.T) {
	type fields struct {
		Collection CollectionLayer
	}
	type args struct {
		ctx    context.Context
		filter interface{}
		opts   []option.FindOneOptioner
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   DocumentResultLayer
	}{
		{
			"FindOne()",
			fields{
				&MongoCollection{},
			},
			args{
				context.Background(),
				nil,
				nil,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields.Collection

			if got := c.FindOne(tt.args.ctx, tt.args.filter, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MongoCollection.FindOne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoCollection_FindOneAndUpdate(t *testing.T) {
	type fields struct {
		Collection CollectionLayer
	}
	type args struct {
		ctx    context.Context
		filter interface{}
		update interface{}
		opts   []option.FindOneAndUpdateOptioner
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   reflect.Type
	}{
		{
			"FindOneAndUpdate()",
			fields{
				&MongoCollection{},
			},
			args{
				context.Background(),
				nil,
				nil,
				nil,
			},
			reflect.TypeOf(&MongoDocumentResult{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields.Collection

			if got := c.FindOneAndUpdate(tt.args.ctx, tt.args.filter, tt.args.update, tt.args.opts...); reflect.TypeOf(got) != tt.want {
				t.Errorf("MongoCollection.FindOneAndUpdate() = %v, want %v", reflect.TypeOf(got), tt.want)
			}
		})
	}
}

func TestMongoCollection_FindOneAndDelete(t *testing.T) {
	type fields struct {
		Collection CollectionLayer
	}
	type args struct {
		ctx    context.Context
		filter interface{}
		opts   []option.FindOneAndDeleteOptioner
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   DocumentResultLayer
	}{
		{
			"FindOneAndDelete()",
			fields{
				&MongoCollection{},
			},
			args{
				context.Background(),
				nil,
				nil,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields.Collection

			if got := c.FindOneAndDelete(tt.args.ctx, tt.args.filter, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MongoCollection.FindOneAndDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMongoDocumentResult_Decode(t *testing.T) {

	type fields struct {
		DocumentResult *mongo.DocumentResult
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"Decode()",
			fields{
				&mongo.DocumentResult{},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := MongoDocumentResult{
				DocumentResult: tt.fields.DocumentResult,
			}

			if _, err := d.Decode(); (err != nil) != tt.wantErr {
				t.Errorf("MongoDocumentResult.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMongoClient_Close(t *testing.T) {

	_, host := testServer(t, nil)
	client, _ := mongo.NewClient("mongodb://" + host)

	type fields struct {
		ctx    context.Context
		Client *mongo.Client
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
				client,
			},
			true, // Can't close it because it never started.
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := MongoClient{
				ctx:    tt.fields.ctx,
				Client: tt.fields.Client,
			}
			if err := mc.Close(); (err != nil) != tt.wantErr {
				t.Errorf("MongoClient.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
