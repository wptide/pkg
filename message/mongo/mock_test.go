package mongo

import (
	"github.com/mongodb/mongo-go-driver/core/options"
	"context"
	"github.com/mongodb/mongo-go-driver/mongo"
	"reflect"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"encoding/json"
	"github.com/wptide/pkg/message"
	wrapper "github.com/wptide/pkg/wrapper/mongo"
)

type MockClient struct {
	collection string
}

func (m MockClient) Database(string) wrapper.DataLayer {
	return &MockDatabase{
		collection: m.collection,
	}
}

func (m MockClient) Close() error {
	return nil
}

type MockDatabase struct {
	collection string
}

func (m MockDatabase) Collection(name string) wrapper.CollectionLayer {
	return &MockCollection{
		collection: m.collection,
	}
}

type MockCollection struct {
	collection string
}

func (m MockCollection) InsertOne(ctx context.Context, document interface{}, opts ...options.InsertOneOptioner) (wrapper.InsertOneResultLayer, error) {
	return nil, nil
}

func (m MockCollection) FindOne(ctx context.Context, filter interface{}, opts ...options.FindOneOptioner) wrapper.DocumentResultLayer {

	switch m.collection {
	case "test-no-records":
		return &wrapper.MongoDocumentResult{
			&mongo.DocumentResult{},
		}
	default:
		return &MockDocumentResult{
			collection: m.collection,
		}
	}
}
func (m MockCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...options.FindOneAndUpdateOptioner) wrapper.DocumentResultLayer {
	return &MockDocumentResult{
		collection: m.collection + "-update",
	}
}
func (m MockCollection) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...options.FindOneAndDeleteOptioner) wrapper.DocumentResultLayer {
	return nil
}

type MockDocumentResult struct {
	collection string
}

func (d MockDocumentResult) Decode(v interface{}) error {
	switch d.collection {

	case "test-valid-message-update":
		fallthrough
	case "test-valid-message":

		msgJson, _ := json.Marshal(&message.Message{
			Title: "Plugin One",
		})
		var msg map[string]interface{}
		json.Unmarshal(msgJson, &msg)

		id, _ := objectid.FromHex("abcdef123456789009876364")
		obj := map[string]interface{}{
			"_id":             id,
			"retry_available": true,
			"lock":            int64(0),
			"created":         int64(0),
			"retries":         int64(3),
			"message":         msg,
		}
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf(obj))

	case "test-valid-message-no-retry-update":
		fallthrough
	case "test-valid-message-no-retry":
		msgJson, _ := json.Marshal(&message.Message{
			Title: "Plugin One",
		})
		var msg map[string]interface{}
		json.Unmarshal(msgJson, &msg)

		id, _ := objectid.FromHex("abcdef123456789009876364")
		obj := map[string]interface{}{
			"_id":             id,
			"retry_available": true,
			"lock":            int64(0),
			"created":         int64(0),
			"retries":         int64(0),
			"message":         msg,
		}
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf(obj))

	case "test-lock-fail-update":
		obj := "failed!"
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf(obj))
	case "test-lock-fail":
		msgJson, _ := json.Marshal(&message.Message{
			Title: "Plugin One",
		})
		var msg map[string]interface{}
		json.Unmarshal(msgJson, &msg)

		id, _ := objectid.FromHex("abcdef123456789009876364")
		obj := map[string]interface{}{
			"_id":             id,
			"retry_available": true,
			"lock":            int64(0),
			"created":         int64(0),
			"retries":         int64(0),
			"message":         msg,
		}
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf(obj))

	case "test-message-fail-update":
		fallthrough
	case "test-message-fail":
		id, _ := objectid.FromHex("abcdef123456789009876364")
		obj := map[string]interface{}{
			"_id":             id,
			"retry_available": true,
			"lock":            int64(0),
			"created":         int64(0),
			"retries":         int64(0),
			"message":         false,
		}
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf(obj))

	}

	return nil
}
