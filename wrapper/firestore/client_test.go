package firestore

import (
	"context"
	"reflect"
	"testing"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go"
	"google.golang.org/api/option"
)

func TestClient_getDocData(t *testing.T) {
	type args struct {
		ss *firestore.DocumentSnapshot
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			"Get Doc Data",
			args{
				&firestore.DocumentSnapshot{},
			},
			nil,
		},
		{
			"Nil Doc Data",
			args{
				nil,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{}
			if got := c.getDocData(tt.args.ss); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.getDocData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetDoc(t *testing.T) {

	mockBase, _ := firebase.NewApp(context.Background(), nil,
		option.WithEndpoint("ws://localhost:5555"),
		option.WithCredentialsFile("./testdata/service-account.json"),
	)
	mockStore, _ := mockBase.Firestore(context.Background())

	type fields struct {
		Firestore *firestore.Client
		Ctx       context.Context
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]interface{}
	}{
		{
			"GetDoc",
			fields{
				mockStore,
				context.Background(),
			},
			args{
				"test-doc",
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				Firestore: tt.fields.Firestore,
				Ctx:       tt.fields.Ctx,
			}
			if got := c.GetDoc(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetDoc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_SetDoc(t *testing.T) {

	mockBase, _ := firebase.NewApp(context.Background(), nil,
		option.WithEndpoint("ws://localhost:5555"),
		option.WithCredentialsFile("./testdata/service-account.json"),
	)
	mockStore, _ := mockBase.Firestore(context.Background())

	type fields struct {
		Firestore *firestore.Client
		Ctx       context.Context
	}
	type args struct {
		path string
		data map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Set Doc Data",
			fields{
				mockStore,
				context.Background(),
			},
			args{
				"test",
				map[string]interface{}{
					"test": "data",
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				Firestore: tt.fields.Firestore,
				Ctx:       tt.fields.Ctx,
			}
			if err := c.SetDoc(tt.args.path, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Client.SetDoc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_AddDoc(t *testing.T) {

	mockBase, _ := firebase.NewApp(context.Background(), nil,
		option.WithEndpoint("ws://localhost:5555"),
		option.WithCredentialsFile("./testdata/service-account.json"),
	)
	mockStore, _ := mockBase.Firestore(context.Background())

	type fields struct {
		Firestore *firestore.Client
		Ctx       context.Context
	}
	type args struct {
		collection string
		data       interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Add Doc",
			fields{
				mockStore,
				context.Background(),
			},
			args{
				"test-collection",
				"test data",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				Firestore: tt.fields.Firestore,
				Ctx:       tt.fields.Ctx,
			}
			if err := c.AddDoc(tt.args.collection, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Client.AddDoc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Close(t *testing.T) {

	mockBase, _ := firebase.NewApp(context.Background(), nil,
		option.WithEndpoint("ws://localhost:5555"),
		option.WithCredentialsFile("./testdata/service-account.json"),
	)
	mockStore, _ := mockBase.Firestore(context.Background())

	type fields struct {
		Firestore *firestore.Client
		Ctx       context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"Close",
			fields{
				mockStore,
				context.Background(),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				Firestore: tt.fields.Firestore,
				Ctx:       tt.fields.Ctx,
			}
			if err := c.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Client.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Authenticated(t *testing.T) {

	mockBase, _ := firebase.NewApp(context.Background(), nil,
		option.WithEndpoint("ws://localhost:5555"),
		option.WithCredentialsFile("./testdata/service-account.json"),
	)
	mockStore, _ := mockBase.Firestore(context.Background())

	type fields struct {
		Firestore *firestore.Client
		Ctx       context.Context
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"Authenticated",
			fields{
				mockStore,
				context.Background(),
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				Firestore: tt.fields.Firestore,
				Ctx:       tt.fields.Ctx,
			}
			if got := c.Authenticated(); got != tt.want {
				t.Errorf("Client.Authenticated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeleteDoc(t *testing.T) {

	mockBase, _ := firebase.NewApp(context.Background(), nil,
		option.WithEndpoint("ws://localhost:5555"),
		option.WithCredentialsFile("./testdata/service-account.json"),
	)
	mockStore, _ := mockBase.Firestore(context.Background())

	type fields struct {
		Firestore *firestore.Client
		Ctx       context.Context
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Delete Doc",
			fields{
				mockStore,
				context.Background(),
			},
			args{
				"delete-doc",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				Firestore: tt.fields.Firestore,
				Ctx:       tt.fields.Ctx,
			}
			if err := c.DeleteDoc(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Client.DeleteDoc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
