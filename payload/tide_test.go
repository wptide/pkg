package payload

import (
	"reflect"
	"testing"

	"github.com/wptide/pkg/message"
	"github.com/wptide/pkg/tide"
	"errors"
)

type MockTideClient struct {
	apiError bool
}

func (m MockTideClient) Authenticate(clientId, clientSecret, authEndpoint string) error {
	return nil
}

func (m MockTideClient) SendPayload(method, endpoint, data string) (string, error) {

	if m.apiError {
		return "", errors.New("API error")
	}

	return "", nil
}

func TestTidePayload_BuildPayload(t *testing.T) {
	type fields struct {
		Client tide.ClientInterface
	}
	type args struct {
		msg  message.Message
		data map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := TidePayload{
				Client: tt.fields.Client,
			}
			got, err := tp.BuildPayload(tt.args.msg, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("TidePayload.BuildPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TidePayload.BuildPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fallbackValue(t *testing.T) {
	type args struct {
		value []interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fallbackValue(tt.args.value...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fallbackValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTidePayload_SendPayload(t *testing.T) {
	type fields struct {
		Client tide.ClientInterface
	}
	type args struct {
		destination string
		payload     []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := TidePayload{
				Client: tt.fields.Client,
			}
			got, err := tp.SendPayload(tt.args.destination, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("TidePayload.SendPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TidePayload.SendPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}
