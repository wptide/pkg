package mongo

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/core/option"
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

func (m MockCollection) InsertOne(ctx context.Context, document interface{}, opts ...option.InsertOneOptioner) (wrapper.InsertOneResultLayer, error) {
	return nil, nil
}

func (m MockCollection) FindOne(ctx context.Context, filter interface{}, opts ...option.FindOneOptioner) wrapper.DocumentResultLayer {

	switch m.collection {
	case "test-no-records":
		return &MockDocumentResult{}
	default:
		return &MockDocumentResult{
			collection: m.collection,
		}
	}
}
func (m MockCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...option.FindOneAndUpdateOptioner) wrapper.DocumentResultLayer {
	return &MockDocumentResult{
		collection: m.collection + "-update",
	}
}
func (m MockCollection) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...option.FindOneAndDeleteOptioner) wrapper.DocumentResultLayer {
	return nil
}

type MockDocumentResult struct {
	collection string
}

func (d MockDocumentResult) Decode() (*bson.Document, error) {

	var document *bson.Document

	switch d.collection {

	case "test-valid-message-update":
		fallthrough
	case "test-valid-message":
		msg := generateMessage(&message.Message{
			Title: "Plugin One",
		})
		msgJson, _ := json.Marshal(msg)

		doc, err := bson.ParseExtJSONObject(string(msgJson))
		id, _ := objectid.FromHex("abcdef123456789009876364")
		doc.Append(bson.EC.ObjectID("_id", id))
		return doc, err

	case "test-valid-message-no-retry-update":
		fallthrough
	case "test-valid-message-no-retry":
		msg := generateMessage(&message.Message{
			Title: "Plugin One",
		})
		msg["retries"] = int64(0)
		msgJson, _ := json.Marshal(msg)

		doc, err := bson.ParseExtJSONObject(string(msgJson))
		id, _ := objectid.FromHex("abcdef123456789009876364")
		doc.Append(bson.EC.ObjectID("_id", id))
		return doc, err

	case "test-lock-fail-update":
		return nil, errors.New("something went wrong")

	case "test-lock-fail":
		msg := generateMessage(&message.Message{
			Title: "Plugin One",
		})
		msgJson, _ := json.Marshal(msg)

		doc, err := bson.ParseExtJSONObject(string(msgJson))
		id, _ := objectid.FromHex("abcdef123456789009876364")
		doc.Append(bson.EC.ObjectID("_id", id))
		return doc, err

	}

	return document, nil
}
