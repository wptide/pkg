package firestore

import (
	"reflect"
	"testing"

	"cloud.google.com/go/firestore"
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
